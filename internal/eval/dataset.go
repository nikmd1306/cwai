//go:build eval

package eval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type Sample struct {
	ID              string   `json:"id"`
	Description     string   `json:"description"`
	Diff            string   `json:"diff"`
	ExpectedMessage string   `json:"expected_message"`
	ExpectedType    string   `json:"expected_type"`
	ExpectedScope   string   `json:"expected_scope"`
	Tags            []string `json:"tags,omitempty"`
}

func LoadDataset(path string) ([]Sample, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open dataset: %w", err)
	}
	defer func() { _ = f.Close() }()

	var samples []Sample
	reader := bufio.NewReader(f)

	lineNum := 0
	for {
		line, err := reader.ReadString('\n')
		lineNum++
		line = strings.TrimRight(line, "\r\n")
		if line != "" {
			var s Sample
			if jsonErr := json.Unmarshal([]byte(line), &s); jsonErr != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, jsonErr)
			}
			samples = append(samples, s)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read dataset: %w", err)
		}
	}
	return samples, nil
}

func LoadDatasetFiltered(path string, tags []string) ([]Sample, error) {
	all, err := LoadDataset(path)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return all, nil
	}

	tagSet := make(map[string]bool, len(tags))
	for _, t := range tags {
		tagSet[t] = true
	}

	var filtered []Sample
	for _, s := range all {
		for _, t := range s.Tags {
			if tagSet[t] {
				filtered = append(filtered, s)
				break
			}
		}
	}
	return filtered, nil
}
