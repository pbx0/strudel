package strudel

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/coreos/strudel/config"
	"github.com/coreos/strudel/update"
)

// Run strudel with valid config from main
func RunTryUpdate(cfg config.AppConfig) error {
	log.Printf("updating pod: %v", cfg.ServiceName)

	initialVersion, err := cfg.GetLocalVersion()
	if err != nil {
		return err
	}

	if initialVersion == "" {
		log.Println("service file not found, one will be created on succesful update")
	}

	poller := update.NewPoller(initialVersion, cfg)
	resp, err := poller.Check()
	if err != nil {
		return fmt.Errorf("error, upate check failed: %v", err)
	}

	if resp == nil {
		log.Println("no update needed")
		//TODO: signal this with error type
		return nil
	}

	updatePayload, err := poller.Download(resp)
	if err != nil {
		log.Println("error: %b", err)
		return err
	}

	// parse json payload
	p, err := NewPayload(updatePayload.Body)
	if err != nil {
		return fmt.Errorf("error parsing payload: %v", err)
	}

	// verify signature
	if p.VerifySig(cfg.PubKeys) != true || cfg.InsecureSkipVerify {
		return fmt.Errorf("unable to verify json signature")
	}

	// TODO: Support multiple update strategies and use coreos/go-systemd. For
	// now this is used just to test basic usage of everything.

	// naive service update -- daemon-reload and then restart
	err = p.OverwriteServiceFile(cfg.GetPath())
	if err != nil {
		return fmt.Errorf("overwriting service file: %v", err)
	}

	cmd := exec.Command("systemctl", "daemon-reload", cfg.ServiceName)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %v", err, b)
	}
	cmd = exec.Command("systemctl", "restart", cfg.ServiceName)
	b, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %v", err, b)
	}

	return nil
}
