package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func libraryTemplateSelection(catalog []wmdesign.LibraryTemplate, family, inline, file string) ([]string, map[string]bool, error) {
	if inline != "" && file != "" {
		return nil, nil, fmt.Errorf("choose either --template-keys or --template-keys-file")
	}
	var keys []string
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, nil, fmt.Errorf("read --template-keys-file %s: %w", file, err)
		}
		keys, err = parseLibraryTemplateKeyFile(data)
		if err != nil {
			return nil, nil, fmt.Errorf("--template-keys-file %s: %w", file, err)
		}
	} else if inline != "" {
		keys = strings.Split(inline, ",")
	}
	if len(keys) > 0 && family != "" {
		return nil, nil, fmt.Errorf("choose --family or --template-keys/--template-keys-file")
	}

	selected := make([]string, 0, len(keys))
	seen := make(map[string]bool, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || seen[key] {
			return nil, nil, fmt.Errorf("invalid or duplicate template key: %s", key)
		}
		found := false
		for _, template := range catalog {
			if template.Key == key {
				found = true
				break
			}
		}
		if !found {
			return nil, nil, fmt.Errorf("unknown template key: %s", key)
		}
		seen[key] = true
		selected = append(selected, key)
	}
	return selected, seen, nil
}

func parseLibraryTemplateKeyFile(data []byte) ([]string, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, fmt.Errorf("file is empty; provide at least one template key")
	}
	if strings.HasPrefix(trimmed, "[") {
		var keys []string
		if err := json.Unmarshal([]byte(trimmed), &keys); err != nil {
			return nil, fmt.Errorf("expected a JSON string array or newline/CSV keys: %w", err)
		}
		if len(keys) == 0 {
			return nil, fmt.Errorf("file contains no template keys")
		}
		return keys, nil
	}

	var keys []string
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		keys = append(keys, strings.Split(line, ",")...)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("file contains no template keys")
	}
	return keys, nil
}
