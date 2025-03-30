package record

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"

	"github.com/AfazTech/b9m/parser"
	"github.com/AfazTech/b9m/servicemanager"
	"github.com/AfazTech/b9m/utils"
)

type DNSRecord struct {
	Name  string
	TTL   int
	Type  RecordType
	Value string
}

type RecordType string

const (
	A     RecordType = "A"
	AAAA  RecordType = "AAAA"
	CNAME RecordType = "CNAME"
	TXT   RecordType = "TXT"
	MX    RecordType = "MX"
	NS    RecordType = "NS"
	PTR   RecordType = "PTR"
)

func AddRecord(domain string, recordType RecordType, sub, value string, ttl int) error {
	if err := utils.ValidateDomain(domain); err != nil {
		return fmt.Errorf("failed to add record to domain %s: %w", domain, err)
	}
	if err := utils.ValidateSubdomain(sub); err != nil {
		return fmt.Errorf("failed to add record to domain %s, invalid subdomain %s: %w", domain, sub, err)
	}
	exists, err := utils.DomainExists(domain)
	if err != nil {
		return fmt.Errorf("error checking existence of domain %s: %w", domain, err)
	}
	if !exists {
		return fmt.Errorf("domain does not exist: %s", domain)
	}
	if ttl <= 0 {
		return fmt.Errorf("invalid TTL %d for domain %s: TTL must be greater than 0", ttl, domain)
	}
	validRecordTypes := []RecordType{A, AAAA, CNAME, TXT, MX, NS, PTR}
	if !slices.Contains(validRecordTypes, recordType) {
		return fmt.Errorf("invalid record type %s for domain %s", recordType, domain)
	}
	if recordType == A || recordType == AAAA {
		if err := utils.ValidateIP(value); err != nil {
			return fmt.Errorf("invalid IP address for record in domain %s: %w", domain, err)
		}
	}
	domains, err := parser.GetDomains()
	if err != nil {
		return fmt.Errorf("failed to retrieve domains for adding record to %s: %w", domain, err)
	}
	zoneFile, exists := domains[domain]
	if !exists {
		return fmt.Errorf("zone file not found for domain: %s", domain)
	}

	rec := DNSRecord{
		Name:  sub,
		TTL:   ttl,
		Type:  recordType,
		Value: value,
	}
	var fullName string
	if rec.Name == "@" {
		fullName = rec.Name
	} else {
		fullName = rec.Name + "." + domain + "."
	}

	recordLine := fmt.Sprintf("%s %d IN %s %s\n", fullName, rec.TTL, rec.Type, rec.Value)

	f, err := os.OpenFile(zoneFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(recordLine)
	if err != nil {
		return fmt.Errorf("failed to add record to domain %s: %w", domain, err)
	}
	return servicemanager.ReloadBind()
}

func DeleteRecord(domain, sub string, rType RecordType, value string) error {
	if err := utils.ValidateDomain(domain); err != nil {
		return fmt.Errorf("failed to delete record from domain %s: %w", domain, err)
	}
	if err := utils.ValidateSubdomain(sub); err != nil {
		return fmt.Errorf("failed to delete record from domain %s, invalid subdomain %s: %w", domain, sub, err)
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
		return fmt.Errorf("failed to retrieve domains for deleting record from %s: %w", domain, err)
	}
	zoneFile, exists := domains[domain]
	if !exists {
		return fmt.Errorf("zone file not found for domain: %s", domain)
	}
	data, err := os.ReadFile(zoneFile)
	if err != nil {
		return err
	}

	pattern := fmt.Sprintf(`(?m)^\s*%s\s+(?:\d+\s+)?(?:\S+\s+)?%s\s+%s\s*$`, regexp.QuoteMeta(sub), regexp.QuoteMeta(string(rType)), regexp.QuoteMeta(value))
	re := regexp.MustCompile(pattern)

	newData := re.ReplaceAllString(string(data), "")

	err = os.WriteFile(zoneFile, []byte(newData), 0644)
	if err != nil {
		return fmt.Errorf("failed to delete record from domain %s: %w", domain, err)
	}
	return servicemanager.ReloadBind()
}

func GetAllRecords(domain string) ([]DNSRecord, error) {
	domains, err := parser.GetDomains()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve domains for getting records of %s: %w", domain, err)
	}
	zoneFile, ok := domains[domain]
	if !ok {
		return nil, fmt.Errorf("domain not found: %s", domain)
	}
	zoneData, err := parser.ParseZoneFile(zoneFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse zone file %s for domain %s: %w", zoneFile, domain, err)
	}
	var records []DNSRecord
	for _, rec := range zoneData.Records {
		var valStr string
		switch v := rec.Value.(type) {
		case string:
			valStr = v
		default:
			b, err := json.Marshal(v)
			if err != nil {
				valStr = fmt.Sprintf("%v", v)
			} else {
				valStr = string(b)
			}
		}
		records = append(records, DNSRecord{
			Name:  rec.Name,
			TTL:   rec.TTL,
			Type:  RecordType(rec.Type),
			Value: valStr,
		})
	}
	return records, nil
}
