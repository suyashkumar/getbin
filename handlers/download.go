package handlers

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
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

// OS is an enum representing an operating system variant
type OS string

// OS enumerated values
const (
	OSWindows = "WINDOWS"
	OSDarwin  = "DARWIN"
	OSLinux   = "LINUX"
	OSEmpty   = ""
)

func (o OS) isValid() bool {
	return o == OSWindows || o == OSDarwin || o == OSLinux
}

// DownloadOptions represents various options that can be supplied to the download endpoint
type DownloadOptions struct {
	OS         OS
	Uncompress bool
}

// Simple regex for now, used both for User-Agent and for matching GitHub release asset names
var isDarwin = regexp.MustCompile(`(?i).*(darwin|macintosh).*`)
var isLinux = regexp.MustCompile(`(?i).*linux.*`)
var isWindows = regexp.MustCompile(`(?i).*windows.*`)
var isX86AMD64 = regexp.MustCompile(`(?i).*(x86|amd64).*`)

var osToTester = map[OS]*regexp.Regexp{
	OSDarwin:  isDarwin,
	OSLinux:   isLinux,
	OSWindows: isWindows,
}

// Download handles resolving the latest GitHub release for the given request and either redirecting the download request
// to that URL or unpacking the binary and writing it into the response if specified
func Download(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	opts := parseDownloadOptions(r.URL)
	rp := releases.GithubRepo{
		Username: ps.ByName("username"),
		Repo:     ps.ByName("repo"),
	}
	log.Printf("New Download request with opts: %+v", opts)
	log.Printf("GitHub Repo: %+v", rp)

	rls, err := releases.Get(rp)
	if err != nil {
		sendErrorWithCode(w, "Unable to get latest release from GitHub", 500)
		log.Println("Unable to get latest release from GitHub", err)
		return
	}
	if len(rls) == 0 {
		sendErrorWithCode(w, "No GitHub Releases for this repo.", 400)
		log.Println("No GitHub releases for this repo.", err)
		return
	}

	latestRelease := rls[0]

	var currentPlatformTest *regexp.Regexp
	if opts.OS != OSEmpty {
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
		if currentPlatformTest.MatchString(file) && isX86AMD64.MatchString(file) { // match os and x86/amd64
			currentAsset = &a
			break
		}
	}

	log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))
	log.Printf("Selected Asset: %s", currentAsset.DownloadURL)
	if !opts.Uncompress {
		w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(currentAsset.DownloadURL)))
		http.Redirect(w, r, currentAsset.DownloadURL, http.StatusMovedPermanently)
	} else {
		// Attempt to uncompress the GitHub release asset
		// Note some assumptions below
		w.Header().Add("Content-Type", "application/octet-stream")
		binaryFile, err := http.Get(currentAsset.DownloadURL)

		if currentAsset.ContentType == releases.ContentTypeTARGZ {
			// assume tar.gz
			if err != nil {
				sendErrorWithCode(
					w,
					fmt.Sprintf("Issue downloading binary from GitHub: %s", currentAsset.DownloadURL),
					500,
				)
				log.Println("Issue with downloading binary")
				log.Println(err)
				return
			}
			// untar and copy, TODO: this currently assumes one file, and assumes tar.gz
			zr, err := gzip.NewReader(binaryFile.Body)
			tr := tar.NewReader(zr)
			if err != nil {
				sendErrorWithCode(w, "Issue uncompressing", 500)
				log.Println("Issue uncompressing", err)
				return
			}

			h, err := tr.Next()
			if err != nil {
				sendErrorWithCode(w, "Issue uncompressing (tar Next)", 500)
				log.Println("Issue uncompressing (tar Next)", err)
				return
			}
			w.Header().Add("Content-Length", strconv.Itoa(int(h.Size)))
			w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", "ssl-proxy"))
			_, err = io.Copy(w, tr)
			if err != nil {
				log.Println(err)
			}
		} else if currentAsset.ContentType == releases.ContentTypeZIP {
			// assume .zip, need to read in whole file to unzip, assume single file
			// TODO: consider size limits in future...
			b, err := ioutil.ReadAll(binaryFile.Body)
			if err != nil {
				log.Println("ERROR: issue reading binary file during zip")
			}
			binaryReader := bytes.NewReader(b)
			zr, err := zip.NewReader(binaryReader, binaryReader.Size())
			if err != nil {
				sendErrorWithCode(w, "Error uncompressing", 500)
			}
			if len(zr.File) == 0 {
				// error condition
				sendErrorWithCode(w, "Only one file in ZIP? Possible corruption", 500)
			}
			file := zr.File[0]
			fileRC, err := file.Open()
			if err != nil {
				sendErrorWithCode(w, "Error uncompressing file", 500)
				log.Println("Error opening file")
			}
			log.Printf("Uncompressed %s", file.Name)
			w.Header().Add("Content-Length", strconv.Itoa(int(file.UncompressedSize64)))
			w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Name))
			_, err = io.Copy(w, fileRC)
			if err != nil {
				log.Println(err)
			}
		}
	}

}

func parseDownloadOptions(u *url.URL) *DownloadOptions {
	opts := DownloadOptions{}
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
