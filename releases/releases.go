package releases

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const GITHUB_API_BASE = "https://api.github.com"

const CONTENT_TYPE_TAR_GZ = "application/x-gzip"
const CONTENT_TYPE_ZIP = "application/zip"

type Asset struct {
	DownloadURL string `json:"browser_download_url"`
	ContentType string `json:"content_type"`
}

type Release struct {
	URL     string  `json:"url"`
	Assets  []Asset `json:"assets"`
	TagName string  `json:"tag_name"`
}

type GithubRepo struct {
	Username string
	Repo     string
}

// Get fetches the releases for a provided GithubRepo
func Get(r GithubRepo) ([]Release, error) {
	resp, err := http.Get(fmt.Sprintf("%s/repos/%s/%s/releases", GITHUB_API_BASE, r.Username, r.Repo))
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
