package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/suyashkumar/getbin/releases"
)

var errStopRedirect = errors.New("an error to stop following redirects")
var defaultGithubAPIResponse = []byte(`[{ 
	  "assets": [
		   {"browser_download_url": "http://localhost/some-file-darwin-x86.tar.gz", "content_type": "application/x-gzip"},
		   {"browser_download_url": "http://localhost/some-file-windows-x86.tar.gz", "content_type": "application/x-gzip"},
		   {"browser_download_url": "http://localhost/some-file-linux-x86.tar.gz", "content_type": "application/x-gzip"},
		   {"browser_download_url": "http://localhost/some-file-linux-arm64.tar.gz", "content_type": "application/x-gzip"},
		   {"browser_download_url": "http://localhost/some-file-windows-arm64.tar.gz", "content_type": "application/x-gzip"},
		   {"browser_download_url": "http://localhost/some-file-darwin-arm64.tar.gz", "content_type": "application/x-gzip"}
	  ]
 }]`)

func TestDownload_Redirect(t *testing.T) {
	cases := []struct {
		name              string
		requestPath       string
		userAgent         string
		githubAPIResponse []byte
		wantRedirect      string
	}{
		{
			name:              "darwin option",
			requestPath:       "/username/repo?os=darwin",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-darwin-x86.tar.gz",
		},
		{
			name:              "darwin user-agent",
			requestPath:       "/username/repo",
			userAgent:         "darwin user agent",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-darwin-x86.tar.gz",
		},
		{
			name:              "darwin option arm64",
			requestPath:       "/username/repo?os=darwin&arch=arm64",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-darwin-arm64.tar.gz",
		},
		{
			name:              "darwin user-agent arm64",
			requestPath:       "/username/repo",
			userAgent:         "darwin user agent arm64",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-darwin-arm64.tar.gz",
		},
		{
			name:              "linux option",
			requestPath:       "/username/repo?os=linux",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-linux-x86.tar.gz",
		},
		{
			name:              "linux user-agent",
			requestPath:       "/username/repo",
			userAgent:         "linux user agent",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-linux-x86.tar.gz",
		},
		{
			name:              "linux option arm64",
			requestPath:       "/username/repo?os=linux&arch=arm64",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-linux-arm64.tar.gz",
		},
		{
			name:              "linux user-agent arm64",
			requestPath:       "/username/repo",
			userAgent:         "linux user agent arm64",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-linux-arm64.tar.gz",
		},
		{
			name:              "windows option",
			requestPath:       "/username/repo?os=windows",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-windows-x86.tar.gz",
		},
		{
			name:              "windows user-agent",
			requestPath:       "/username/repo",
			userAgent:         "windows user agent",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-windows-x86.tar.gz",
		},
		{
			name:              "windows option arm64",
			requestPath:       "/username/repo?os=windows&arch=arm64",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-windows-arm64.tar.gz",
		},
		{
			name:              "windows user-agent arm64",
			requestPath:       "/username/repo",
			userAgent:         "windows user agent arm64",
			githubAPIResponse: defaultGithubAPIResponse,
			wantRedirect:      "http://localhost/some-file-windows-arm64.tar.gz",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			wantRequestURL := "/repos/username/repo/releases"

			// Setup fake GitHub server.
			githubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.String() != wantRequestURL {
					t.Errorf("unexpected request URL. got: %v, want: %v", req.URL.String(), wantRequestURL)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Write(tc.githubAPIResponse)
			}))
			defer githubServer.Close()
			tempSetGithubAPIBase(t, githubServer.URL)

			// Setup test getbin server to wrap the Download handler.
			router := httprouter.New()
			router.GET("/:username/:repo", Download)
			server := httptest.NewServer(router)
			defer server.Close()

			// Make test request and client to getbin.
			req, err := http.NewRequest(http.MethodGet, server.URL+tc.requestPath, nil)
			if err != nil {
				t.Fatalf("unable to make GET request to getbin server: %v", err)
			}
			req.Header.Set("User-Agent", tc.userAgent)

			cl := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Need to return an error to stop the client from auto
				// redirecting, since we want to inspect the redirect.
				return errStopRedirect
			}}

			res, err := cl.Do(req)
			if !errors.Is(err, errStopRedirect) {
				t.Errorf("Unexpected error when making getbin request: %v", err)
			}

			if res.StatusCode != http.StatusMovedPermanently {
				t.Errorf("unexpected StatusCode in response. got: %v, want: %v", res.StatusCode, http.StatusMovedPermanently)
			}
			if got := res.Header.Get("Location"); got != tc.wantRedirect {
				t.Errorf("unexpected redirect in response. got: %v, want: %v", got, tc.wantRedirect)
			}
		})
	}
}

func tempSetGithubAPIBase(t *testing.T, newAPIBase string) {
	origVal := releases.GithubAPIBase
	releases.GithubAPIBase = newAPIBase
	t.Cleanup(func() {
		releases.GithubAPIBase = origVal
	})
}
