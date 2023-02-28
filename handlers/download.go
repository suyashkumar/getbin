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
	"github.com/suyashkumar/getbin/releases"
)

// Simple regex for now, used both for User-Agent and for matching GitHub release asset names
var isDarwin = regexp.MustCompile(`(?i).*(darwin|macintosh).*`)
var isLinux = regexp.MustCompile(`(?i).*linux.*`)
var isWindows = regexp.MustCompile(`(?i).*windows.*`)
var isX86AMD64 = regexp.MustCompile(`(?i).*(x86|amd64).*`)
var isARM64 = regexp.MustCompile(`(?i).*arm64.*`)

var osToRegexp = map[OS]*regexp.Regexp{
	OSDarwin:  isDarwin,
	OSLinux:   isLinux,
	OSWindows: isWindows,
}

var archToRegexp = map[Arch]*regexp.Regexp{
	ArchX86:   isX86AMD64,
	ArchAMD64: isX86AMD64,
	ArchARM64: isARM64,
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

	userAgent := r.Header.Get("User-Agent")
	selectedOSRegexp := getOSRegexp(opts, userAgent)
	selectedArchRegexp := getArchRegexp(opts, userAgent)

	var currentAsset *releases.Asset
	for _, a := range latestRelease.Assets {
		file := path.Base(a.DownloadURL)
		if selectedOSRegexp.MatchString(file) && selectedArchRegexp.MatchString(file) {
			currentAsset = &a
			break
		}
	}

	if currentAsset == nil {
		sendErrorWithCode(w, "Unable to find a release asset for your platform in the latest release.", 400)
		return
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

// getOSRegexp returns the appropriate OS identifying regexp based on the
// downloadOptions and (if needed) the userAgent.
func getOSRegexp(opts *downloadOptions, userAgent string) *regexp.Regexp {
	if opts.OS != OSEmpty {
		return osToRegexp[opts.OS]
	}

	// Check userAgent to infer OS:
	for _, osRegexp := range osToRegexp {
		if osRegexp.MatchString(userAgent) {
			return osRegexp
		}
	}

	return isLinux // Note: Linux is the default
}

// getArchRegexp returns the appropriate Arch identifying regexp based on the
// downloadOptions and (if needed) the userAgent.
func getArchRegexp(opts *downloadOptions, userAgent string) *regexp.Regexp {
	if opts.Arch != ArchEmpty {
		return archToRegexp[opts.Arch]
	}

	// Check userAgent to infer Arch:
	for _, archRegexp := range archToRegexp {
		if archRegexp.MatchString(userAgent) {
			return archRegexp
		}
	}

	return isX86AMD64 // Note: x86 / amd64 is the default.
}

// downloadOptions represents various options that can be supplied to the download endpoint
type downloadOptions struct {
	OS         OS
	Arch       Arch
	Uncompress bool
}

func parseDownloadOptions(u *url.URL) *downloadOptions {
	opts := downloadOptions{}
	if val, ok := u.Query()["os"]; ok {
		os := OS(strings.ToUpper(val[0]))
		if os.isValid() {
			opts.OS = os
		}
	}

	if val, ok := u.Query()["arch"]; ok {
		arch := Arch(strings.ToUpper(val[0]))
		if arch.isValid() {
			opts.Arch = arch
			log.Println("opts", opts.Arch)
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

// OS is an enum representing an operating system variant.
type OS string

// OS enumerated values
const (
	OSWindows = OS("WINDOWS")
	OSDarwin  = OS("DARWIN")
	OSLinux   = OS("LINUX")
	OSEmpty   = OS("")
)

func (o OS) isValid() bool {
	return o == OSWindows || o == OSDarwin || o == OSLinux
}

// Arch is an enum representing supported architectures.
type Arch string

const (
	ArchAMD64 = Arch("AMD64")
	ArchX86   = Arch("X86")
	ArchARM64 = Arch("ARM64")
	ArchEmpty = Arch("")
)

func (a Arch) isValid() bool {
	return a == ArchAMD64 || a == ArchX86 || a == ArchARM64
}
