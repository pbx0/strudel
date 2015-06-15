package update

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/coreos/go-omaha/omaha"
	"github.com/coreos/strudel/config"
)

type Poller struct {
	currentVersion string
	appConfig      config.AppConfig
}

func NewPoller(version string, appConfig config.AppConfig) *Poller {
	return &Poller{version, appConfig}
}

func (p *Poller) Check() (*updateCheckResponse, error) {
	resp, err := p.sendEvent(omahaEventComplete)
	if resp == nil || err != nil {
		return nil, err
	}

	if resp.Status == UpdateCheckStatusNoUpdate || resp.Status == UpdateCheckStatusErrorVersion {
		return nil, nil
	}

	if resp.Version == p.currentVersion {
		return nil, nil
	}

	return resp, nil
}

func (p *Poller) Download(resp *updateCheckResponse) (*OmahaPayload, error) {
	p.sendEventAndLog(omahaEventDownloading)

	body, err := fetchURL(resp.PackageURL)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch payload of package v%s from %s: %v", resp.Version, resp.PackageURL, err)
	}

	pl := &OmahaPayload{
		Version: resp.Version,
		Body:    body,
	}

	p.sendEventAndLog(omahaEventDownloaded)
	p.sendEventAndLog(omahaEventInstalled)

	return pl, nil
}

// Send an Omaha request to the update service.
func (p *Poller) sendEvent(event omahaEvent) (*updateCheckResponse, error) {
	client := &http.Client{}

	request := omaha.NewRequest("", "", "", "")
	app := request.AddApp(p.appConfig.AppID, p.currentVersion)
	app.AddUpdateCheck()
	app.Track = p.appConfig.Group

	e := app.AddEvent()
	e.Type = strconv.Itoa(event.Type)
	e.Result = strconv.Itoa(event.Result)

	raw, err := xml.MarshalIndent(request, "", " ")
	if err != nil {
		return nil, fmt.Errorf("failed marshaling omaha request: %v", err)
	}

	// log.Debugf("Sending omaha request: %s%s", xml.Header, raw)

	resp, err := client.Post(p.appConfig.OmahaEndpoint(), "text/xml", bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("failed sending omaha event to server: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal omaha response: %v", err)
	}

	// log.Debugf("omaha response: %s%s", xml.Header, string(body))

	var oresp omaha.Response
	if err = xml.Unmarshal(body, &oresp); err != nil {
		return nil, err
	}

	return parseUpdateCheckResponse(oresp.Apps[0].UpdateCheck)
}

// Just send a status and disregard the response.
func (p *Poller) sendEventAndLog(event omahaEvent) {
	if _, err := p.sendEvent(event); err != nil {
		log.Printf("Error sending %q event: %v", event.Name, err)
	}
}

// Parse all needed data out of the update check's response.
func parseUpdateCheckResponse(uc *omaha.UpdateCheck) (*updateCheckResponse, error) {
	// Don't parse out package details if they dont exist.
	if uc.Status == UpdateCheckStatusNoUpdate || uc.Status == UpdateCheckStatusErrorVersion {
		return nil, nil
	}

	u, err := url.Parse(uc.Urls.Urls[0].CodeBase)
	if err != nil {
		return nil, fmt.Errorf("failed parsing url: %v", err)
	}

	name := uc.Manifest.Packages.Packages[0].Name
	u.Path = path.Join(u.Path, name)

	resp := updateCheckResponse{
		Status:      uc.Status,
		PackageURL:  u.String(),
		PackageName: name,
		Version:     uc.Manifest.Version,
	}

	return &resp, nil
}
