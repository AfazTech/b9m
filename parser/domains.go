package parser

import (
	"fmt"
	"strings"

	"github.com/AfazTech/b9m/config"
)

func GetDomains() (map[string]string, error) {
	configFile := config.GetConfigFile()
	zoneDir := config.GetZoneDir()
	config, err := ParseConfig(configFile)

	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration file %s: %w", configFile, err)
	}

	domains := make(map[string]string)
	for key, val := range config {
		if strings.HasPrefix(key, "zone") {
			var domain string
			if strings.HasPrefix(key, "zone \"") {
				parts := strings.SplitN(key, "\"", 3)
				if len(parts) >= 2 {
					domain = parts[1]
				}
			} else {
				parts := strings.Fields(key)
				if len(parts) >= 2 {
					domain = parts[1]
				}
			}
			if domain != "" {
				if zoneMap, ok := val.(map[string]interface{}); ok {
					if fileVal, ok2 := zoneMap["file"]; ok2 {
						filePath, _ := fileVal.(string)
						if !strings.HasPrefix(filePath, "/") {
							filePath = zoneDir + "/" + filePath
						}
						domains[domain] = filePath
					}
				}
			}
		}
	}
	return domains, nil
}
