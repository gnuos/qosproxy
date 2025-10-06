package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

const RULE_FILE_NAME = "rules.yml"

var cfg *Config

type Config struct {
	Rules []string `yaml:"rules"`
}

func ParseConfig(p string) (*Config, error) {
	var c Config

	f, err := os.Open(p)
	if err != nil {
		log.Fatalf("打开规则文件%s失败\n", p)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(content, &c)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)

		return nil, err
	}

	return &c, nil
}

func searchConfig() (err error) {
	var cwd, userHome string

	cwd, err = os.Getwd()
	if err != nil {
		return
	}

	if cfgPath != "" {
		if fileReadable(cfgPath) {
			cfgPath, err = filepath.Abs(cfgPath)
			if err != nil {
				cfgPath = ""
				return
			}

			return
		}

		cfgPath = ""
	} else {
		cfgPath = fmt.Sprintf("%s/etc/%s", cwd, RULE_FILE_NAME)
		if fileReadable(cfgPath) {
			return
		}

		userHome, err = os.UserHomeDir()
		if err != nil {
			cfgPath = ""
			return
		}

		cfgPath = fmt.Sprintf("%s/.%s", userHome, RULE_FILE_NAME)
		if dirExists(userHome) && fileReadable(cfgPath) {
			return
		}

		cfgPath = ""
	}

	return
}

func fileReadable(f string) bool {
	info, err := os.Stat(f)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	if info.Mode().Perm()&0444 != 0444 {
		return false
	}

	return true
}

func dirExists(d string) bool {
	info, err := os.Stat(d)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	if !info.IsDir() {
		return false
	}

	return true
}
