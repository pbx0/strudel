package update

import (
	"io/ioutil"
	"net/http"
)

const (
	UpdateCheckStatusErrorVersion = "error-version"
	UpdateCheckStatusNoUpdate     = "noupdate"
)

var (
	omahaEventDownloading = omahaEvent{Name: "downloading", Type: 13, Result: 1}
	omahaEventDownloaded  = omahaEvent{Name: "downloaded", Type: 14, Result: 1}
	omahaEventInstalled   = omahaEvent{Name: "installed", Type: 3, Result: 1}
	omahaEventComplete    = omahaEvent{Name: "complete", Type: 3, Result: 2}
	omahaEventHold        = omahaEvent{Name: "hold", Type: 800, Result: 1}
	omahaEventError       = omahaEvent{Name: "error", Type: 3, Result: 0}
)

type omahaEvent struct {
	Name   string
	Type   int
	Result int
}

// Relevant information in response from an update check.
type updateCheckResponse struct {
	Version     string
	PackageURL  string
	PackageName string
	Status      string
}

type OmahaPayload struct {
	Version string
	Body    []byte
}

func fetchURL(loc string) ([]byte, error) {
	req, err := http.NewRequest("GET", loc, nil)
	if err != nil {
		return nil, err
	}

	c := &http.Client{
		Transport: http.DefaultTransport,
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	return ioutil.ReadAll(resp.Body)
}
