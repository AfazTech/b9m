package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AfazTech/logger/v2"
)

var confcommentRegex = regexp.MustCompile(`^\s*(//.*|#.*)?$`)
var keyValueRegex = regexp.MustCompile(`^\s*(\S+)\s+(.+?)\s*(;)?$`)
var blockStartRegex = regexp.MustCompile(`^\s*(\S.*?)\s*{\s*$`)
var inlineBlockRegex = regexp.MustCompile(`^\s*(\S+)\s*{\s*(.*?)\s*}\s*(;)?$`)
var includeRegex = regexp.MustCompile(`^\s*include\s+["']?([^"']+)["']?\s*(;)?$`)

func parseBlock(scanner *bufio.Scanner) map[string]interface{} {
	block := make(map[string]interface{})
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if confcommentRegex.MatchString(line) {
			continue
		}
		if line == "}" || line == "};" {
			return block
		}
		if blockStartMatches := blockStartRegex.FindStringSubmatch(line); blockStartMatches != nil {
			block[blockStartMatches[1]] = parseBlock(scanner)
			continue
		}
		if inlineMatches := inlineBlockRegex.FindStringSubmatch(line); inlineMatches != nil {
			key := inlineMatches[1]
			listContent := inlineMatches[2]
			parts := strings.Split(listContent, ";")
			var values []string
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					values = append(values, part)
				}
			}
			if len(values) == 1 {
				block[key] = values[0]
			} else {
				block[key] = values
			}
			continue
		}
		if kvMatches := keyValueRegex.FindStringSubmatch(line); kvMatches != nil {
			block[kvMatches[1]] = strings.Trim(kvMatches[2], "\"")
		}
	}
	return block
}

func mergeConfig(dst, src map[string]interface{}) {
	for k, v := range src {
		if existing, ok := dst[k]; ok {
			if dstMap, ok1 := existing.(map[string]interface{}); ok1 {
				if srcMap, ok2 := v.(map[string]interface{}); ok2 {
					mergeConfig(dstMap, srcMap)
					continue
				}
			}
		}
		dst[k] = v
	}
}

func ParseConfig(file string) (map[string]interface{}, error) {
	config := make(map[string]interface{})
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	baseDir := filepath.Dir(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if confcommentRegex.MatchString(line) {
			continue
		}
		if includeMatches := includeRegex.FindStringSubmatch(line); includeMatches != nil {
			includePath := includeMatches[1]
			if !filepath.IsAbs(includePath) {
				includePath = filepath.Join(baseDir, includePath)
			}
			incConfig, err := ParseConfig(includePath)
			if err == nil {
				mergeConfig(config, incConfig)
			}
			continue
		}
		if blockStartMatches := blockStartRegex.FindStringSubmatch(line); blockStartMatches != nil {
			config[blockStartMatches[1]] = parseBlock(scanner)
			continue
		}
		if inlineMatches := inlineBlockRegex.FindStringSubmatch(line); inlineMatches != nil {
			key := inlineMatches[1]
			listContent := inlineMatches[2]
			parts := strings.Split(listContent, ";")
			var values []string
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					values = append(values, part)
				}
			}
			if len(values) == 1 {
				config[key] = values[0]
			} else {
				config[key] = values
			}
			continue
		}
		if kvMatches := keyValueRegex.FindStringSubmatch(line); kvMatches != nil {
			config[kvMatches[1]] = strings.Trim(kvMatches[2], "\"")
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}
	return config, nil
}
