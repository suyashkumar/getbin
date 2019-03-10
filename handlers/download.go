package handlers

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/suyashkumar/bin/releases"
)

type OS string
const OS_WINDOWS = "WINDOWS"
const OS_DARWIN = "DARWIN"
const OS_LINUX = "LINUX"
const OS_EMPTY = ""
func (o OS) isValid() bool {
	return o == OS_WINDOWS || o == OS_DARWIN || o == OS_LINUX
}

type DownloadOptions struct {
	OS OS
	Uncompress bool
}

// Simple regex for now, used both for User-Agent and for matching GitHub release asset names
var isDarwin = regexp.MustCompile(`(?i).*darwin.*`)
var isLinux = regexp.MustCompile(`(?i).*linux.*`)
var isWindows = regexp.MustCompile(`(?i).*windows.*`)

var osToTester = map[OS]*regexp.Regexp{
	OS_DARWIN: isDarwin,
	OS_LINUX: isLinux,
	OS_WINDOWS: isWindows,
}

// Download handles resolving the latest GitHub release for the given request and either redirecting the download request
// to that URL or unpacking the binary and writing it into the response if specified
func Download(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	opts := parseDownloadOptions(r.URL)

	rls, err := releases.Get(releases.GithubRepo{
		Username: ps.ByName("username"),
		Repo: ps.ByName("repo"),
	})

	if err != nil {

	}
	latestRelease := rls[0]

	var currentPlatformTest *regexp.Regexp
	if opts.OS != OS_EMPTY {
		currentPlatformTest = osToTester[opts.OS]
	} else {
		currentPlatformTest = isLinux // Note: linux is the default
		userAgent := r.Header.Get("User-Agent")
		for _, isOS := range osToTester {
			if isOS.MatchString(userAgent) {
				currentPlatformTest = isOS
				break
			}
		}
	}

	var currentAsset *releases.Asset
	for _, a := range latestRelease.Assets {
		file := path.Base(a.DownloadURL)
		if currentPlatformTest.MatchString(file) {
			currentAsset = &a
			break
		}
	}
	log.Println(r.Header.Get("User-Agent"))

	if !opts.Uncompress {
		w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(currentAsset.DownloadURL)))
		http.Redirect(w, r, currentAsset.DownloadURL, http.StatusMovedPermanently)
	} else {
		// Attempt to uncompress the GitHub release asset
		// Note some assumptions below
		if currentAsset.ContentType == releases.CONTENT_TYPE_TAR_GZ {
			w.Header().Add("Content-Type", "application/octet-stream")
			binaryFile, err := http.Get(currentAsset.DownloadURL)
			if err != nil {
				log.Println("Issue with downloading binary")
				log.Println(err)
			}
			// untar and copy, TODO: this currently assumes one file, and assumes tar.gz
			zr, err := gzip.NewReader(binaryFile.Body)
			tr := tar.NewReader(zr)
			if err != nil {
				log.Println(err)
			}
			h, err := tr.Next()
			w.Header().Add("Content-Length", strconv.Itoa(int(h.Size)))
			w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", "ssl-proxy"))
			written, err := io.Copy(w, tr)
			if err != nil {
				log.Println(err)
			}
			log.Println(written)
		}
	}

}

func parseDownloadOptions(u *url.URL) *DownloadOptions {
	opts := DownloadOptions{}
	log.Println(u.Query()["os"])
	if val, ok := u.Query()["os"]; ok {
		os := OS(strings.ToUpper(val[0]))
		if os.isValid() {
			opts.OS = os
		}
	}

	if val, ok := u.Query()["uncompress"]; ok {
		uncompress, err := strconv.ParseBool(val[0])
		if err == nil {
			opts.Uncompress = uncompress
		}
	}

	return &opts

}