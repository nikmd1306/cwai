package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	checkInterval = 24 * time.Hour
	requestTimeout = 5 * time.Second
	stateFileName  = ".cwai.state"
	releaseURL     = "https://api.github.com/repos/nikmd1306/cwai/releases/latest"
)

type StateEntry struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
	LatestURL     string    `json:"latest_url"`
}

type ReleaseInfo struct {
	Version string
	URL     string
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func stateFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, stateFileName), nil
}

func readState() (*StateEntry, error) {
	path, err := stateFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &StateEntry{}, nil
		}
		return nil, err
	}
	var state StateEntry
	if err := json.Unmarshal(data, &state); err != nil {
		return &StateEntry{}, nil
	}
	return &state, nil
}

func writeState(state *StateEntry) error {
	path, err := stateFilePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func CheckForUpdate(currentVersion string) (*ReleaseInfo, error) {
	if currentVersion == "dev" {
		return nil, nil
	}

	state, err := readState()
	if err != nil {
		return nil, err
	}

	if time.Since(state.CheckedAt) < checkInterval {
		if state.LatestVersion != "" && compareVersions(state.LatestVersion, currentVersion) > 0 {
			return &ReleaseInfo{Version: state.LatestVersion, URL: state.LatestURL}, nil
		}
		return nil, nil
	}

	client := &http.Client{Timeout: requestTimeout}
	resp, err := client.Get(releaseURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")

	state.CheckedAt = time.Now()
	state.LatestVersion = latestVersion
	state.LatestURL = release.HTMLURL
	_ = writeState(state)

	if compareVersions(latestVersion, currentVersion) > 0 {
		return &ReleaseInfo{Version: latestVersion, URL: release.HTMLURL}, nil
	}

	return nil, nil
}

func compareVersions(a, b string) int {
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	for i := 0; i < 3; i++ {
		var numA, numB int
		if i < len(partsA) {
			_, _ = fmt.Sscanf(partsA[i], "%d", &numA)
		}
		if i < len(partsB) {
			_, _ = fmt.Sscanf(partsB[i], "%d", &numB)
		}
		if numA > numB {
			return 1
		}
		if numA < numB {
			return -1
		}
	}
	return 0
}
