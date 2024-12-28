package controller

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type BindManager struct {
	zoneDir string
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

func NewBindManager(zoneDir string) *BindManager {
	return &BindManager{zoneDir: zoneDir}
}

func (bm *BindManager) validateDomain(domain string) error {
	if matched, _ := regexp.MatchString(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`, domain); !matched {
		return errors.New("invalid domain format")
	}
	return nil
}

func (bm *BindManager) domainExists(domain string) (bool, error) {
	zoneFile := fmt.Sprintf("%s/db.%s", bm.zoneDir, domain)
	_, err := os.Stat(zoneFile)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (bm *BindManager) AddDomain(domain string) error {
	if err := bm.validateDomain(domain); err != nil {
		return err
	}
	exists, err := bm.domainExists(domain)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("domain already exists")
	}
	zoneFile := fmt.Sprintf("%s/db.%s", bm.zoneDir, domain)
	record := fmt.Sprintf("$TTL 86400\n@ IN SOA ns1.%s. admin.%s. ( 2023100101 3600 1800 604800 86400 )\n", domain, domain)
	return bm.createZoneFile(zoneFile, record)
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
	zoneFile := fmt.Sprintf("%s/db.%s", bm.zoneDir, domain)
	return os.Remove(zoneFile)
}

func (bm *BindManager) AddRecord(domain string, recordType RecordType, name, value string, ttl int) error {
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
	if ttl <= 0 {
		return errors.New("TTL must be greater than 0")
	}
	validRecordTypes := []RecordType{A, CNAME, TXT, MX, NS, PTR}
	if !contains(validRecordTypes, recordType) {
		return errors.New("invalid record type")
	}
	zoneFile := fmt.Sprintf("%s/db.%s", bm.zoneDir, domain)
	record := fmt.Sprintf("%s IN %s %d %s", name, recordType, ttl, value)
	return bm.addRecord(zoneFile, record)
}

func (bm *BindManager) DeleteRecord(domain, name string) error {
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
	zoneFile := fmt.Sprintf("%s/db.%s", bm.zoneDir, domain)
	return bm.deleteRecord(zoneFile, name+".")
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
	return nil
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
	return nil
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
	return os.WriteFile(zoneFile, []byte(strings.Join(newLines, "\n")), 0644)
}

func contains(slice []RecordType, item RecordType) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
