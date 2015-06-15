package config

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/strudel/auth"
)

const serviceExt = ".service.conf"

type AppConfig struct {
	ServiceName        string
	AppID              string
	Endpoint           string
	Group              string
	UnitDir            string
	UpdateMethod       string
	InsecureSkipVerify bool
	PubKeys            auth.PubKeys

	localVersion string //sha256 hash of service file
}

func (c *AppConfig) Valid() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service name not set")
	}
	if c.AppID == "" {
		return fmt.Errorf("app ID not set")
	}
	if c.Endpoint == "" {
		return fmt.Errorf("omaha endpoint not set")
	}
	if c.Group == "" {
		return fmt.Errorf("group not set")
	}
	if c.UnitDir == "" {
		return fmt.Errorf("unit-dir not set")
	}
	if len(c.PubKeys) == 0 && c.InsecureSkipVerify == false {
		return fmt.Errorf("no public keys specified while insecure-skip-verify is false")
	}
	if len(c.PubKeys) > 0 && c.InsecureSkipVerify == true {
		return fmt.Errorf("pubic keys specified while insecure-skip-verify is set")
	}
	return nil
}

func (cfg *AppConfig) OmahaEndpoint() string {
	return fmt.Sprintf("%s/v1/update/", cfg.Endpoint)
}

// GetLocalVersion gets the version hash of the service file. If the file
// doesn't exist, an empty string will be returned.
// TODO: add semantic versioning support
func (cfg *AppConfig) GetLocalVersion() (string, error) {
	// TODO: Open Question: should the user be warned about updating a pod for
	// which other service files exist? If there was a RO service file it could
	// conflict with future upgrades to the given service file.

	// locate expected service file
	filebytes, err := ioutil.ReadFile(cfg.GetPath())
	if err == os.ErrNotExist {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading service file: %v", err)
	}

	cfg.localVersion = getVersionHash(filebytes)
	return cfg.localVersion, nil
}

func (cfg *AppConfig) GetPath() string {
	return filepath.Join(cfg.UnitDir, cfg.ServiceName+serviceExt)
}

func getVersionHash(filebytes []byte) string {
	hasher := sha256.New()
	hasher.Write(filebytes)
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}
