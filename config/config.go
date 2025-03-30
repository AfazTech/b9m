package config

import (
	"github.com/AfazTech/logger/v2"
)

var (
	confFile string
	zoneDir  string
	err      error
)

func init() {
	confFile, err = detectConfigFile()
	if err != nil {
		logger.Fatalf("failed to detect config file: %v", err)
	}
	zoneDir, err = detectZoneDir()
	if err != nil {
		logger.Fatalf("failed to detect zone directory: %v", err)
	}
}

func GetConfigFile() string {
	return confFile
}

func GetZoneDir() string {
	return zoneDir
}
