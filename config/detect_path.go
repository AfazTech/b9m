package config

import (
	"errors"
	"os"
)

func detectConfigFile() (string, error) {
	configPaths := []string{
		"/etc/named.conf",
		"/etc/bind/named.conf",
		"/usr/local/etc/named.conf",
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New("config file not found")
}

func detectZoneDir() (string, error) {
	zoneDirs := []string{
		"/var/named",
		"/var/lib/bind",
		"/etc/bind",
		"/usr/local/etc/namedb",
	}

	for _, dir := range zoneDirs {
		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			return dir, nil
		}
	}

	return "", errors.New("zone directory not found")
}
