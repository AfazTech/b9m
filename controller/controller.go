package controller

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type BindManager struct {
	zoneDir       string
	namedConfFile string
	client        *dns.Client
}

type DNSRecord struct {
	Name  string
	TTL   int
	Type  RecordType
	Value string
}

type RecordType string

const (
	A     RecordType = "A"
	CNAME RecordType = "CNAME"
	TXT   RecordType = "TXT"
	MX    RecordType = "MX"
	NS    RecordType = "NS"
	PTR   RecordType = "PTR"
)

func NewBindManager(zoneDir string, namedConfFile string) *BindManager {
	return &BindManager{
		zoneDir:       zoneDir,
		namedConfFile: namedConfFile,
		client:        new(dns.Client),
	}
}

func (bm *BindManager) validateSubdomain(sub string) error {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9-]{1,63}$`, sub)
	if !matched {
		return fmt.Errorf("invalid subdomain format: %s", sub)
	}
	return nil
}

func (bm *BindManager) validateDomain(domain string) error {
	if matched, _ := regexp.MatchString(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`, domain); !matched {
		return fmt.Errorf("invalid domain format: %s", domain)
	}
	return nil
}

func (bm *BindManager) validateARecord(nsName string) error {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(nsName), dns.TypeA)
	m.RecursionDesired = true

	_, _, err := bm.client.Exchange(m, "8.8.8.8:53")
	if err != nil {
		return fmt.Errorf("failed to resolve A record for %s: %w", nsName, err)
	}
	return nil
}

func (bm *BindManager) validateIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return errors.New("invalid IP address format")
	}
	return nil
}

func (bm *BindManager) domainExists(domain string) (bool, error) {
	domains, err := bm.GetDomains()
	if err != nil {
		return false, err
	}
	_, exists := domains[domain]
	return exists, nil
}

func (bm *BindManager) AddDomain(domain string, ns1 string, ns2 string) error {
	if err := bm.validateDomain(domain); err != nil {
		return err
	}
	if err := bm.validateDomain(ns1); err != nil {
		return err
	}
	if err := bm.validateDomain(ns2); err != nil {
		return err
	}

	exists, err := bm.domainExists(domain)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("domain already exists")
	}

	if err := bm.validateARecord(ns1); err != nil {
		return err
	}

	if err := bm.validateARecord(ns2); err != nil {
		return err
	}

	zoneFile := fmt.Sprintf("%s/%s.b9m", bm.zoneDir, domain)
	record := fmt.Sprintf("$TTL 86400\n@ IN SOA %s. admin.%s. ( 2023100101 86400 3600 604800 86400 )\n", ns1, domain)
	record += fmt.Sprintf("@ IN NS %s.\n", ns1)
	record += fmt.Sprintf("@ IN NS %s.\n", ns2)

	if err := bm.createZoneFile(zoneFile, record); err != nil {
		return err
	}
	return bm.addZone(domain)
}

func (bm *BindManager) recordExists(zoneFile, sub string) (bool, error) {
	data, err := os.ReadFile(zoneFile)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("%s.", sub)) {
			return true, nil
		}
	}
	return false, nil
}

func (bm *BindManager) DeleteDomain(domain string) error {
	if err := bm.validateDomain(domain); err != nil {
		return err
	}
	exists, err := bm.domainExists(domain)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("domain does not exist")
	}

	domains, err := bm.GetDomains()
	if err != nil {
		return err
	}
	zoneFile, exists := domains[domain]
	if !exists {
		return errors.New("domain file not found")
	}

	if err := os.Remove(zoneFile); err != nil {
		return err
	}
	return bm.deleteZone(domain)
}

func (bm *BindManager) AddRecord(domain string, recordType RecordType, sub, value string, ttl int) error {
	if err := bm.validateDomain(domain); err != nil {
		return err
	}
	if err := bm.validateSubdomain(sub); err != nil {
		return err
	}
	exists, err := bm.domainExists(domain)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("domain does not exist")
	}
	if ttl <= 0 {
		return errors.New("TTL must be greater than 0")
	}
	validRecordTypes := []RecordType{A, CNAME, TXT, MX, NS, PTR}
	if !contains(validRecordTypes, recordType) {
		return errors.New("invalid record type")
	}
	if recordType == A {
		if err := bm.validateIP(value); err != nil {
			return err
		}
	}

	domains, err := bm.GetDomains()
	if err != nil {
		return err
	}
	zoneFile, exists := domains[domain]
	if !exists {
		return errors.New("domain file not found")
	}

	recordExists, err := bm.recordExists(zoneFile, sub)
	if err != nil {
		return err
	}
	if recordExists {
		return fmt.Errorf("record with subdomain %s already exists", sub)
	}

	record := fmt.Sprintf("%s.%s. %d IN %s %s", sub, domain, ttl, recordType, value)
	return bm.addRecord(zoneFile, record)
}

func (bm *BindManager) DeleteRecord(domain, sub string) error {
	if err := bm.validateDomain(domain); err != nil {
		return err
	}
	if err := bm.validateSubdomain(sub); err != nil {
		return err
	}
	exists, err := bm.domainExists(domain)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("domain does not exist")
	}

	domains, err := bm.GetDomains()
	if err != nil {
		return err
	}
	zoneFile, exists := domains[domain]
	if !exists {
		return errors.New("domain file not found")
	}

	return bm.deleteRecord(zoneFile, fmt.Sprintf("%s.%s.", sub, domain))
}

func (bm *BindManager) createZoneFile(zoneFile, record string) error {
	file, err := os.Create(zoneFile)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(record); err != nil {
		return err
	}
	return bm.reloadBind()
}

func (bm *BindManager) addRecord(zoneFile, record string) error {
	file, err := os.OpenFile(zoneFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(record + "\n"); err != nil {
		return err
	}
	return bm.reloadBind()
}

func (bm *BindManager) deleteRecord(zoneFile, name string) error {
	data, err := os.ReadFile(zoneFile)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.Contains(line, name) {
			newLines = append(newLines, line)
		}
	}
	if err := os.WriteFile(zoneFile, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return err
	}
	return bm.reloadBind()
}

func contains(slice []RecordType, item RecordType) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func (bm *BindManager) addZone(domain string) error {
	zoneEntry := fmt.Sprintf("zone \"%s\" {\n\ttype master;\n\tfile \"%s/%s.b9m\";\n};\n", domain, bm.zoneDir, domain)

	data, err := os.ReadFile(bm.namedConfFile)
	if err != nil {
		return err
	}
	if strings.Contains(string(data), domain) {
		return errors.New("zone already exists in named.conf.local")
	}

	file, err := os.OpenFile(bm.namedConfFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(zoneEntry); err != nil {
		return err
	}

	return bm.reloadBind()
}

func (bm *BindManager) deleteZone(domain string) error {
	data, err := os.ReadFile(bm.namedConfFile)
	if err != nil {
		return err
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

	if err := os.WriteFile(bm.namedConfFile, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return err
	}
	return bm.reloadBind()
}

func (bm *BindManager) reloadBind() error {
	cmd := exec.Command("rndc", "reload")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (bm *BindManager) GetAllRecords(domain string) ([]DNSRecord, error) {
	if err := bm.validateDomain(domain); err != nil {
		return nil, err
	}

	exists, err := bm.domainExists(domain)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("domain does not exist")
	}

	domains, err := bm.GetDomains()
	if err != nil {
		return nil, err
	}
	zoneFile, exists := domains[domain]
	if !exists {
		return nil, errors.New("domain file not found")
	}

	data, err := os.ReadFile(zoneFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var records []DNSRecord

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "$") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}

		name := parts[0]

		ttl := 3600
		if _, err := strconv.Atoi(parts[1]); err == nil {
			ttl, _ = strconv.Atoi(parts[1])
			parts = parts[1:]
		}

		recordType := RecordType(parts[1])
		value := strings.Join(parts[2:], " ")

		if !strings.HasSuffix(name, domain+".") && !strings.HasSuffix(name, ".") {
			name = name + "." + domain + "."
		}

		if !strings.Contains(line, "IN") {
			value = "IN " + value
		} else {
			value = strings.Replace(value, "IN ", "", 1)
		}

		record := DNSRecord{
			Name:  name,
			TTL:   ttl,
			Type:  recordType,
			Value: value,
		}
		records = append(records, record)
	}

	return records, nil
}

func (bm *BindManager) ReloadBind() error {
	cmd := exec.Command("rndc", "reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to Reload Bind: %v", err)
	}
	return nil
}

func (bm *BindManager) RestartBind() error {
	cmd := exec.Command("systemctl", "restart", "bind9")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to restart Bind: %v", err)
	}
	return nil
}

func (bm *BindManager) StopBind() error {
	cmd := exec.Command("systemctl", "stop", "bind9")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to stop bind: %v", err)
	}
	return nil
}

func (bm *BindManager) StartBind() error {
	cmd := exec.Command("systemctl", "start", "bind9")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to start Bind: %v", err)
	}
	return nil
}

func (bm *BindManager) StatusBind() (string, error) {
	cmd := exec.Command("systemctl", "is-active", "bind9")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get bind status: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

type Stats struct {
	TotalZones   int `json:"total_zones"`
	TotalRouters int `json:"total_routers"`
}

func (bm *BindManager) GetStats() (Stats, error) {
	zones, err := ioutil.ReadDir(bm.zoneDir)
	if err != nil {
		return Stats{}, errors.New("failed to read zone directory")
	}
	namedConfData, err := ioutil.ReadFile(bm.namedConfFile)
	if err != nil {
		return Stats{}, errors.New("failed to read named.conf.local")
	}
	routerCount := strings.Count(string(namedConfData), "zone ")

	return Stats{
		TotalZones:   len(zones),
		TotalRouters: routerCount,
	}, nil
}

func (bm *BindManager) GetDomains() (map[string]string, error) {
	data, err := os.ReadFile(bm.namedConfFile)
	if err != nil {
		return nil, err
	}

	domains := make(map[string]string)
	lines := strings.Split(string(data), "\n")

	var currentDomain string
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "zone ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentDomain = strings.Trim(parts[1], "\"")
			}
		}
		if strings.Contains(line, "file") && currentDomain != "" {
			parts := strings.Fields(line)
			for i := 0; i < len(parts); i++ {
				if parts[i] == "file" && i+1 < len(parts) {
					filePath := strings.Trim(parts[i+1], "\";")
					domains[currentDomain] = filePath
					currentDomain = ""
					break
				}
			}
		}
	}

	return domains, nil
}
