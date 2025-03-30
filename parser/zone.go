package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type SOARecord struct {
	MName   string
	RName   string
	Serial  int
	Refresh int
	Retry   int
	Expire  int
	Minimum int
}

type ZoneRecord struct {
	Name  string
	TTL   int
	Class string
	Type  string
	Value interface{}
}

type ZoneData struct {
	Origin  string
	TTL     int
	Records []ZoneRecord
}

var recordRegex = regexp.MustCompile(`^(\S+)\s+(\d+)?\s*(IN)?\s*(A|AAAA|CNAME|MX|NS|SOA|TXT|PTR|SRV)\s+(.+)$`)
var soaRegex = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\d+)\s+(\d+[SMHDW]?)\s+(\d+[SMHDW]?)\s+(\d+[SMHDW]?)\s+(\d+[SMHDW]?)$`)
var mxRegex = regexp.MustCompile(`^(\d+)\s+(\S+)$`)
var srvRegex = regexp.MustCompile(`^(\d+)\s+(\d+)\s+(\d+)\s+(\S+)$`)
var originRegex = regexp.MustCompile(`^\$ORIGIN\s+(\S+)$`)
var ttlRegex = regexp.MustCompile(`^\$TTL\s+(\d+[SMHDW]?)$`)
var commentRegex = regexp.MustCompile(`^(.*?)(;.*)?$`)

var timeUnits = map[rune]int{
	'S': 1,
	'M': 60,
	'H': 3600,
	'D': 86400,
	'W': 604800,
}

func parseTTL(ttlStr string) (int, error) {
	if ttl, err := strconv.Atoi(ttlStr); err == nil {
		return ttl, nil
	}
	numStr := ttlStr[:len(ttlStr)-1]
	unit := rune(ttlStr[len(ttlStr)-1])
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, err
	}
	multiplier, ok := timeUnits[unit]
	if !ok {
		return 0, fmt.Errorf("unknown TTL unit")
	}
	return num * multiplier, nil
}

func parseRecord(tokens []string, globalTTL int, globalOrigin string) (ZoneRecord, error) {
	line := strings.Join(tokens, " ")
	matches := recordRegex.FindStringSubmatch(line)
	if matches == nil {
		return ZoneRecord{}, fmt.Errorf("invalid record format")
	}

	record := ZoneRecord{
		Name:  matches[1],
		Class: "IN",
		Type:  matches[4],
	}
	if matches[2] != "" {
		ttl, _ := parseTTL(matches[2])
		record.TTL = ttl
	} else {
		record.TTL = globalTTL
	}
	if matches[3] != "" {
		record.Class = matches[3]
	}
	if !strings.HasSuffix(record.Name, ".") && globalOrigin != "" && record.Name != globalOrigin {
		record.Name += "." + globalOrigin
	}

	rdata := strings.TrimSpace(matches[5])
	switch record.Type {
	case "SOA":
		soaMatches := soaRegex.FindStringSubmatch(rdata)
		if soaMatches == nil {
			return ZoneRecord{}, fmt.Errorf("invalid SOA format")
		}
		serial, _ := strconv.Atoi(soaMatches[3])
		refresh, _ := parseTTL(soaMatches[4])
		retry, _ := parseTTL(soaMatches[5])
		expire, _ := parseTTL(soaMatches[6])
		minimum, _ := parseTTL(soaMatches[7])
		record.Value = SOARecord{
			MName:   soaMatches[1],
			RName:   soaMatches[2],
			Serial:  serial,
			Refresh: refresh,
			Retry:   retry,
			Expire:  expire,
			Minimum: minimum,
		}
	case "NS", "A", "AAAA", "CNAME", "PTR":
		record.Value = rdata
	case "MX":
		mxMatches := mxRegex.FindStringSubmatch(rdata)
		if mxMatches == nil {
			return ZoneRecord{}, fmt.Errorf("invalid MX format")
		}
		pref, _ := strconv.Atoi(mxMatches[1])
		record.Value = map[string]interface{}{
			"preference": pref,
			"exchange":   mxMatches[2],
		}
	case "SRV":
		srvMatches := srvRegex.FindStringSubmatch(rdata)
		if srvMatches == nil {
			return ZoneRecord{}, fmt.Errorf("invalid SRV format")
		}
		priority, _ := strconv.Atoi(srvMatches[1])
		weight, _ := strconv.Atoi(srvMatches[2])
		port, _ := strconv.Atoi(srvMatches[3])
		record.Value = map[string]interface{}{
			"priority": priority,
			"weight":   weight,
			"port":     port,
			"target":   srvMatches[4],
		}
	case "TXT":
		record.Value = strings.Trim(rdata, "\"")
	}
	return record, nil
}

func ParseZoneFile(filePath string) (ZoneData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return ZoneData{}, err
	}
	defer file.Close()

	zone := ZoneData{Records: []ZoneRecord{}}
	scanner := bufio.NewScanner(file)
	multiLineRegex := regexp.MustCompile(`\((.*?)\)`)
	globalOrigin := ""
	globalTTL := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		commentMatches := commentRegex.FindStringSubmatch(line)
		line = strings.TrimSpace(commentMatches[1])
		if line == "" {
			continue
		}

		if multiLineMatches := multiLineRegex.FindStringSubmatch(line); multiLineMatches != nil {
			line = strings.TrimSpace(multiLineMatches[1])
		} else if strings.Contains(line, "(") && !strings.Contains(line, ")") {
			fullLine := line
			for scanner.Scan() {
				nextLine := strings.TrimSpace(scanner.Text())
				if nextLine == "" {
					continue
				}
				commentMatches = commentRegex.FindStringSubmatch(nextLine)
				nextLine = strings.TrimSpace(commentMatches[1])
				fullLine += " " + nextLine
				if strings.Contains(nextLine, ")") {
					break
				}
			}
			line = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(fullLine, "(", ""), ")", ""))
		}

		if originMatches := originRegex.FindStringSubmatch(line); originMatches != nil {
			globalOrigin = originMatches[1]
			if !strings.HasSuffix(globalOrigin, ".") {
				globalOrigin += "."
			}
			continue
		}

		if ttlMatches := ttlRegex.FindStringSubmatch(line); ttlMatches != nil {
			globalTTL, _ = parseTTL(ttlMatches[1])
			continue
		}

		tokens := strings.Fields(line)
		if recordRegex.MatchString(line) {
			rec, err := parseRecord(tokens, globalTTL, globalOrigin)
			if err == nil {
				zone.Records = append(zone.Records, rec)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return zone, err
	}
	zone.Origin = globalOrigin
	zone.TTL = globalTTL

	return zone, nil
}
