package handlers

import (
	"net/http"
	"path"
	"regexp"

	"github.com/julienschmidt/httprouter"
	"github.com/suyashkumar/bin/releases"
)

var isDarwin = regexp.MustCompile(`.*darwin.*`)

func Download(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rls, err := releases.Get(releases.GithubRepo{
		Username: ps.ByName("username"),
		Repo: ps.ByName("repo"),
	})

	if err != nil {

	}

	latestRelease := rls[0]

	// Assume darwin for now
	var currentAsset *releases.Asset
	for _, a := range latestRelease.Assets {
		file := path.Base(a.DownloadURL)
		if isDarwin.MatchString(file) {
			currentAsset = &a
			break
		}
	}

	http.Redirect(w, r, currentAsset.DownloadURL, http.StatusMovedPermanently)

	/*
	if currentAsset.ContentType == releases.CONTENT_TYPE_TAR_GZ {
		w.Header().Add("Content-Type", "application/octet-stream")
		binaryFile, err := http.Get(currentAsset.DownloadURL)
		if err != nil {
			log.Println("Issue with downloading binary")
			log.Println(err)
		}
		// untar and copy, TODO: this currently assumes one file
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
	}*/

}
