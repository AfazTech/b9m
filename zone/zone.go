package zone

import (
	"fmt"
	"os"
	"strings"

	"github.com/AfazTech/b9m/config"
	"github.com/AfazTech/b9m/parser"
	"github.com/AfazTech/b9m/servicemanager"
	"github.com/AfazTech/b9m/utils"
)

func AddDomain(domain string, ns1 string, ns2 string) error {
	if err := utils.ValidateDomain(domain); err != nil {
		return fmt.Errorf("failed to add domain %s: %w", domain, err)
	}
	if err := utils.ValidateDomain(ns1); err != nil {
		return fmt.Errorf("invalid nameserver domain for NS1 (%s): %w", ns1, err)
	}
	if err := utils.ValidateDomain(ns2); err != nil {
		return fmt.Errorf("invalid nameserver domain for NS2 (%s): %w", ns2, err)
	}
	exists, err := utils.DomainExists(domain)
	if err != nil {
		return fmt.Errorf("error checking existence of domain %s: %w", domain, err)
	}
	if exists {
		return fmt.Errorf("domain already exists: %s", domain)
	}
	if err := utils.ValidateARecord(ns1); err != nil {
		return fmt.Errorf("failed to validate A record for NS1 (%s): %w", ns1, err)
	}
	if err := utils.ValidateARecord(ns2); err != nil {
		return fmt.Errorf("failed to validate A record for NS2 (%s): %w", ns2, err)
	}

	zoneFile := fmt.Sprintf("%s/%s.b9m", config.GetZoneDir(), domain)
	record := fmt.Sprintf("$TTL 86400\n@ IN SOA %s. admin.%s. ( 2023100101 86400 3600 604800 86400 )\n", ns1, domain)
	record += fmt.Sprintf("@ IN NS %s.\n", ns1)
	record += fmt.Sprintf("@ IN NS %s.\n", ns2)

	file, err := os.Create(zoneFile)
	if err != nil {
		return fmt.Errorf("failed to create zone file for domain %s: %w", domain, err)
	}
	defer file.Close()
	if _, err := file.WriteString(record); err != nil {
		return fmt.Errorf("failed to write zone record to file %s: %w", zoneFile, err)
	}

	return addZone(domain)
}

func DeleteDomain(domain string) error {
	if err := utils.ValidateDomain(domain); err != nil {
		return fmt.Errorf("failed to delete domain %s: %w", domain, err)
	}
	exists, err := utils.DomainExists(domain)
	if err != nil {
		return fmt.Errorf("error checking existence of domain %s: %w", domain, err)
	}
	if !exists {
		return fmt.Errorf("domain does not exist: %s", domain)
	}
	domains, err := parser.GetDomains()
	if err != nil {
		return fmt.Errorf("failed to retrieve domains for deletion of %s: %w", domain, err)
	}
	zoneFile, exists := domains[domain]
	if !exists {
		return fmt.Errorf("zone file not found for domain: %s", domain)
	}
	if err := os.Remove(zoneFile); err != nil {
		return fmt.Errorf("failed to remove zone file %s for domain %s: %w", zoneFile, domain, err)
	}
	return deleteZone(domain)
}

func deleteZone(domain string) error {
	confFile := config.GetConfigFile()
	data, err := os.ReadFile(confFile)
	if err != nil {
		return fmt.Errorf("failed to read configuration file %s: %w", confFile, err)
	}
	lines := strings.Split(string(data), "\n")
	var newLines []string
	zoneEntry := fmt.Sprintf("zone \"%s\" {", domain)
	skipNextLines := false
	for _, line := range lines {
		if skipNextLines {
			if line == "};" {
				skipNextLines = false
			}
			continue
		}
		if strings.HasPrefix(line, zoneEntry) {
			skipNextLines = true
			continue
		}
		newLines = append(newLines, line)
	}
	if err := os.WriteFile(confFile, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to update configuration file after deleting zone for domain %s: %w", domain, err)
	}
	return servicemanager.ReloadBind()
}

func addZone(domain string) error {
	confFile := config.GetConfigFile()
	zoneDir := config.GetZoneDir()
	zoneEntry := fmt.Sprintf("zone \"%s\" {\n\ttype master;\n\tfile \"%s/%s.b9m\";\n};\n", domain, zoneDir, domain)
	data, err := os.ReadFile(config.GetConfigFile())
	if err != nil {
		return fmt.Errorf("failed to read configuration file %s: %w", confFile, err)
	}
	if strings.Contains(string(data), domain) {
		return fmt.Errorf("zone for domain %s already exists in configuration", domain)
	}
	file, err := os.OpenFile(confFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open configuration file %s for appending: %w", confFile, err)
	}
	defer file.Close()
	if _, err := file.WriteString(zoneEntry); err != nil {
		return fmt.Errorf("failed to write zone entry for domain %s to configuration file: %w", domain, err)
	}
	return servicemanager.ReloadBind()
}
