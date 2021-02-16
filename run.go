package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/sirupsen/logrus"
)

type inventoryHash struct {
	Filename string
	Hash     uint64
}

type metadata struct {
	APIVersion  string
	ProfileHash uint64
	Gzipped     bool
	Inventory   []string
	QueryHashes []inventoryHash
}

func run(config *config, profile *profile, apiVersion string) error {

	// log.WithFields(log.Fields{"num_mccs": len(c.Accounts.MCCs)}).Info("Processing MCCs")
	for i, mcc := range config.Accounts.MCCs {

		mccLog := logrus.WithFields(logrus.Fields{"MCC": mcc.ID})
		mccLog.Infof("MCC %d/%d", i+1, len(config.Accounts.MCCs))

		for i, accID := range mcc.AccountIDs {

			accountLog := mccLog.WithFields(logrus.Fields{"Account": accID})
			accountLog.Infof("Account %d/%d", i+1, len(mcc.AccountIDs))

			if err := processAccount(apiVersion, accID, mcc.ID, config, profile, accountLog); err != nil {
				return fmt.Errorf("process account %q: %w", accID, err)
			}
		}
	}

	for i, accID := range config.Accounts.Direct {

		accountLog := logrus.WithFields(logrus.Fields{"Account": accID})
		accountLog.Infof("Direct Account %d/%d", i+1, len(config.Accounts.Direct))

		if err := processAccount(apiVersion, accID, "", config, profile, accountLog); err != nil {
			return fmt.Errorf("process account %q, %w", accID, err)
		}
	}

	return nil
}

func processAccount(apiVersion string, accountID, mccID string, config *config, profile *profile, accountLog logrus.FieldLogger) error {
	accountOutputDir := fmt.Sprintf("%s/%s", config.OutputDir, accountID)
	err := os.Mkdir(accountOutputDir, os.FileMode(0777))
	if err != nil {
		return fmt.Errorf("create account output dir: %w", err)
	}

	logrus.Infof("Using API version: %s", apiVersion)
	logrus.Infof("Writing download metadata to: %s", accountOutputDir)
	writeMeta(accountOutputDir, apiVersion, profile, gzipOutput)

	for i, q := range profile.Queries {

		queryLog := accountLog.WithFields(logrus.Fields{
			"MCC":     mccID,
			"Account": accountID,
			"Query":   q.FilePrefix,
		})

		queryLog.Infof(fmt.Sprintf("Account Query %d/%d", i+1, len(profile.Queries)))

		var nextToken *string
		pageN := 1

		for pageN == 1 || nextToken != nil {
			pageSize := 10000
			startTime := time.Now()

			accessToken, err := getAccessToken(
				config.Access.ClientID, config.Access.ClientSecret, config.Access.RefreshToken)
			if err != nil {
				return fmt.Errorf("get access token: %w", err)
			}

			page, err := q.execute(
				apiVersion,
				accessToken,
				config.Access.DeveloperToken,
				mccID,
				accountID,
				&pageSize,
				nextToken,
				config.BackoffIntervals,
				queryLog,
			)
			if err != nil {
				return fmt.Errorf("query %q (page %d): %w", q.FilePrefix, pageN, err)
			}

			if err = writePage(accountOutputDir, q.FilePrefix, pageN, page, gzipOutput); err != nil {
				return fmt.Errorf("query %q (page %d): write to disk: %w", q.FilePrefix, pageN, err)
			}

			totalRows, n, err := extractPaginationInfo(page)
			nextToken = n
			if err != nil {
				return fmt.Errorf("query %q (page %d): extract pagination: %w", q.FilePrefix, pageN, err)
			}

			queryLog.Debugf("Page %d (%v)", pageN, time.Since(startTime))

			if nextToken != nil && pageN == 1 {
				totalPages := int(math.Ceil((float64(totalRows) / float64(pageSize))))
				queryLog.Debugf("Estimated total pages: %d", totalPages)
			}

			pageN++
		}
	}

	return nil
}

func extractPaginationInfo(payload []byte) (total int, next *string, err error) {
	var responseInfo struct {
		Total string  `json:"totalResultsCount"`
		Next  *string `json:"nextPageToken"`
	}
	err = json.Unmarshal(payload, &responseInfo)
	if err != nil {
		return 0, nil, fmt.Errorf("unmarshall response: %w", err)
	}

	next = responseInfo.Next

	if responseInfo.Total == "" {
		total = 0
		return
	}

	total, err = strconv.Atoi(responseInfo.Total)
	if err != nil {
		return 0, nil, fmt.Errorf("convert total to int: %w", err)
	}

	return
}

func writeMeta(dir string, apiVersion string, profile *profile, gz bool) error {
	fname := fmt.Sprintf("%s/metadata.json", dir)

	hash, err := hashstructure.Hash(profile, nil)
	if err != nil {
		return fmt.Errorf("Failed to generate profile hash")
	}

	var prefixes []string
	var inventoryHashes []inventoryHash
	for _, q := range profile.Queries {
		prefixes = append(prefixes, q.FilePrefix)
		queryHash, err := hashstructure.Hash(q.Query, nil)
		if err != nil {
			return fmt.Errorf("Failed to generate query hash for %s", q.FilePrefix)
		}
		inventoryHash := inventoryHash{
			Filename: q.FilePrefix,
			Hash:     queryHash,
		}
		inventoryHashes = append(inventoryHashes, inventoryHash)
	}

	metadata := &metadata{
		APIVersion:  apiVersion,
		ProfileHash: hash,
		Inventory:   prefixes,
		Gzipped:     gz,
		QueryHashes: inventoryHashes,
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("Failed to JSON marshal metadata")
	}

	if gz {
		fname = fmt.Sprintf("%s.gz", fname)
		comp, err := gzipCompress(data)
		if err != nil {
			return fmt.Errorf("compress page: :%w", err)
		}
		data = comp
	}

	return ioutil.WriteFile(fname, data, os.FileMode(0777))
}

func writePage(dir, prefix string, pageNumber int, data []byte, gz bool) error {
	fname := fmt.Sprintf("%s/%s__page_%d.json", dir, prefix, pageNumber)

	if gz {
		fname = fmt.Sprintf("%s.gz", fname)
		comp, err := gzipCompress(data)
		if err != nil {
			return fmt.Errorf("compress page: :%w", err)
		}
		data = comp
	}

	return ioutil.WriteFile(fname, data, os.FileMode(0777))
}

func gzipCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}

	if err = gz.Flush(); err != nil {
		return nil, err
	}

	if err = gz.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
