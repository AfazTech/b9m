package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/AfazTech/b9m/config"
	"github.com/AfazTech/b9m/parser"
	"github.com/miekg/dns"
)

func ValidateSubdomain(sub string) error {
	if sub == "@" {
		return nil
	}
	matched, _ := regexp.MatchString("^[a-zA-Z0-9-]{1,63}$", sub)
	if !matched {
		return fmt.Errorf("invalid subdomain format: %s", sub)
	}
	return nil
}

func ValidateDomain(domain string) error {
	if matched, _ := regexp.MatchString("^([a-zA-Z0-9-]+\\.)+[a-zA-Z]{2,}$", domain); !matched {
		return fmt.Errorf("invalid domain format: %s", domain)
	}
	return nil
}

func ValidateARecord(nsName string) error {
	c := &dns.Client{Timeout: 2 * time.Second}
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(nsName), dns.TypeA)
	m.RecursionDesired = true

	msg, _, err := c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		return fmt.Errorf("failed to resolve A record for %s: %w", nsName, err)
	}

	if msg.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("failed to resolve A record for %s: received Rcode %d", nsName, msg.Rcode)
	}

	return nil
}

func ValidateIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}
	return nil
}
func DomainExists(domain string) (bool, error) {
	domains, err := parser.GetDomains()
	if err != nil {
		return false, fmt.Errorf("failed to retrieve domains for existence check on %s: %w", domain, err)
	}
	_, exists := domains[domain]
	return exists, nil
}

func Backup(backupDir string) error {
	confFile := config.GetConfigFile()

	backupConfigPath := filepath.Join(backupDir, confFile)
	if err := os.MkdirAll(filepath.Dir(backupConfigPath), 0755); err != nil {
		return err
	}

	input, err := os.ReadFile(confFile)
	if err != nil {
		return err
	}
	if err := os.WriteFile(backupConfigPath, input, 0644); err != nil {
		return err
	}

	config, err := parser.ParseConfig(confFile)
	if err != nil {
		return err
	}

	for _, v := range config {
		if zone, ok := v.(map[string]interface{}); ok {
			if t, exists := zone["type"]; exists && t == "master" {
				if src, ok := zone["file"].(string); ok {
					backupFilePath := filepath.Join(backupDir, src)
					if err := os.MkdirAll(filepath.Dir(backupFilePath), 0755); err != nil {
						continue
					}
					input, err := os.ReadFile(src)
					if err != nil {
						continue
					}
					os.WriteFile(backupFilePath, input, 0644)
				}
			}

		}
	}
	return nil
}
