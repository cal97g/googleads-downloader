package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

// config contains the target accounts and access information.
type config struct {
	OutputDir        string            `yaml:"output_dir" validate:"required"`
	Access           access            `yaml:"access" validate:"required"`
	BackoffIntervals []int             `yaml:"backoff_intervals"`
	TemplateVars     map[string]string `yaml:"template_vars" validate:"required"`
	Accounts         struct {
		Direct []string `yaml:"direct"`
		MCCs   []struct {
			ID         string   `yaml:"mcc_id" validate:"required"`
			AccountIDs []string `yaml:"account_ids" validate:"required"`
		} `yaml:"mcc"`
	} `yaml:"accounts" validate:"required"`
}

type access struct {
	ClientID       string `yaml:"client_id" validate:"required"`
	ClientSecret   string `yaml:"client_secret" validate:"required"`
	RefreshToken   string `yaml:"refresh_token" validate:"required"`
	DeveloperToken string `yaml:"developer_token" validate:"required"`
}

func (c *config) Validate() error {
	validate := *validator.New()
	return validate.Struct(c)
}

// ParseTemplateVars modifies the config's template vars in place
func (c *config) ParseTemplateVars() error {
	for k, v := range c.TemplateVars {
		v, err := parseTemplateVar(k, v, c.TemplateVars)
		if err != nil {
			return fmt.Errorf("parse template var %s: %w", k, err)
		}
		c.TemplateVars[k] = v
	}

	return nil
}

func configFromPath(path string) (*config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %v", err)
	}

	var conf config
	if err := yaml.Unmarshal(bytes, &conf); err != nil {
		return nil, fmt.Errorf("unmarshall: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %v", err)
	}

	if err := conf.ParseTemplateVars(); err != nil {
		return nil, fmt.Errorf("parse template vars: %w", err)
	}

	return &conf, nil
}

// Matching pattern for dynamic template vars
var dynVarRegex = regexp.MustCompile(`{{(\w*)\(([,\s\w]*)\)}}`)

// Replace dynamic template vars with their parsed values
func parseTemplateVar(k, v string, allVars map[string]string) (string, error) {
	dynVarMatch := dynVarRegex.FindStringSubmatch(v)
	if dynVarMatch == nil {
		return v, nil
	}

	logrus.Infof("parsing dynamic template var: %s - %s", k, v)
	fname, fargs := dynVarMatch[1], dynVarMatch[2]
	parser := parserMap[fname]
	if parser == nil {
		return "", fmt.Errorf("%s not found in parser map", fname)
	}

	// recursive parsing of args
	args := strings.Split(fargs, ",")
	for i, arg := range args {
		arg = strings.TrimSpace(arg)
		args[i] = arg
		if arg != k {
			argValue, ok := allVars[arg]
			if ok {
				narg, err := parseTemplateVar(arg, argValue, allVars)
				if err != nil {
					return "", fmt.Errorf("parse referenced var: %w", err)
				}
				args[i] = narg
			}
		}
	}

	parsed, err := parser(args...)
	if err != nil {
		return "", fmt.Errorf("parse var %s: %w", v, err)
	}

	return parsed, nil
}

// map of functions available to dynamic template vars
var parserMap map[string]func(args ...string) (string, error) = map[string]func(args ...string) (string, error){
	"DateSubDays": func(args ...string) (string, error) {
		if len(args) != 2 {
			return "", fmt.Errorf("incorrect number of args")
		}

		dt, err := time.Parse("2006-01-02", args[0])
		if err != nil {
			return "", err
		}

		days, err := strconv.Atoi(args[1])
		if err != nil {
			return "", err
		}

		return dt.AddDate(0, 0, days*-1).Format("2006-01-02"), nil
	},
	"Concat": func(args ...string) (string, error) {
		out := ""
		for _, arg := range args {
			out += arg
		}
		return out, nil
	},
	"Earliest": func(args ...string) (string, error) {
		if len(args) != 2 {
			return "", fmt.Errorf("incorrect number of args")
		}

		dt1, err := time.Parse("2006-01-02", args[0])
		if err != nil {
			return "", err
		}

		dt2, err := time.Parse("2006-01-02", args[1])
		if err != nil {
			return "", err
		}

    if dt1.Before(dt2) {
      return dt1.Format("2006-01-02"), nil
    }
    return dt2.Format("2006-01-02"), nil
	},
}
