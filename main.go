package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

var gzipOutput = false

func main() {

	var apiVersion string

	configPath := flag.String("config", "", "The path to the input conf file")
	profilePath := flag.String("profile", "", "The path to the profile file")
	flag.StringVar(&apiVersion, "api-version", "", "The version of the Google Ads API to call")
	compressFlag := flag.Bool("compress", false, "Set gzipped output")
	verboseFlag := flag.Bool("verbose", false, "Set verbose output")

	flag.Parse()

	if *verboseFlag {
		log.SetLevel(log.DebugLevel)
	}

	gzipOutput = *compressFlag

	if *configPath == "" {
		logrus.Fatal("missing -config")
		os.Exit(1)
	}

	if *profilePath == "" {
		logrus.Fatal("missing -profile")
		os.Exit(1)
	}

	if apiVersion == "" {
		logrus.Fatal("missing -api-version")
		os.Exit(1)
	}

	config, err := configFromPath(*configPath)
	if err != nil {
		logrus.Errorf("load conf: %v", err)
		os.Exit(1)
	}

	profile, err := profileFromPath(*profilePath)
	if err != nil {
		logrus.Errorf("load profile: %v", err)
		os.Exit(1)
	}

	if err := profile.interpolate(config.TemplateVars); err != nil {
		logrus.Errorf("interpolate template vars: %v", err)
		os.Exit(1)
	}

	err = run(config, profile, apiVersion)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
