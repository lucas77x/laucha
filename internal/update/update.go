// Package update checks GitHub Releases for a newer version.
package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// latestURL is a variable so tests can point it at a fake server.
var latestURL = "https://api.github.com/repos/lucas77x/laucha/releases/latest"

// ReleasePage is where users download new versions.
const ReleasePage = "https://github.com/lucas77x/laucha/releases/latest"

// ErrNoReleases means the repository has no published releases yet.
var ErrNoReleases = errors.New("no releases published")

var client = &http.Client{Timeout: 10 * time.Second}

// CheckLatest returns the latest published version tag.
func CheckLatest() (string, error) {
	resp, err := client.Get(latestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrNoReleases
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api: %s", resp.Status)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", ErrNoReleases
	}
	return release.TagName, nil
}

// IsNewer reports whether latest is a higher semantic version than
// current. Malformed versions compare as not newer.
func IsNewer(current, latest string) bool {
	c, okCurrent := parse(current)
	l, okLatest := parse(latest)
	if !okCurrent || !okLatest {
		return false
	}
	for i := 0; i < 3; i++ {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

func parse(version string) ([3]int, bool) {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	parts := strings.SplitN(version, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}
