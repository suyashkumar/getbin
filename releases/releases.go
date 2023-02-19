package releases

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// GithubAPIBase represents the base url for the GitHub API. This can be
// changed, primarily for test purposes.
var GithubAPIBase = "https://api.github.com"

// Content type constants
const (
	ContentTypeTARGZ = "application/x-gzip"
	ContentTypeZIP   = "application/zip"
)

// Asset represents an uploaded GitHub release asset
type Asset struct {
	DownloadURL string `json:"browser_download_url"`
	ContentType string `json:"content_type"`
}

// Release represents a GitHub release entity
type Release struct {
	URL     string  `json:"url"`
	Assets  []Asset `json:"assets"`
	TagName string  `json:"tag_name"`
}

// GithubRepo represents a unique repository on the GitHub platform
type GithubRepo struct {
	Username string
	Repo     string
}

// Get fetches the releases for a provided GithubRepo
func Get(r GithubRepo) ([]Release, error) {
	resp, err := http.Get(fmt.Sprintf("%s/repos/%s/%s/releases", GithubAPIBase, r.Username, r.Repo))
	if err != nil {
		log.Println("ERROR: issue making API request")
		log.Println(err)
		return nil, err
	}

	dec := json.NewDecoder(resp.Body)
	var releases []Release
	if err = dec.Decode(&releases); err != nil {
		log.Println("ERROR: issue decoding Github API response")
		log.Println(err)
		return nil, err
	}

	return releases, nil
}
