package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/coreos/strudel/config"
	"github.com/coreos/strudel/strudel"
)

const (
	cliName        = "strudel"
	cliDescription = "strudel, upgrade rkt pods with omaha"
)

var (
	globalFlagset = flag.NewFlagSet(cliName, flag.ExitOnError)
)

func main() {
	var cfg config.AppConfig

	fs := flag.NewFlagSet(cliName, flag.ExitOnError)
	fs.StringVar(&cfg.ServiceName, "service-name", "", "systemd service file name to locally identify app")
	fs.StringVar(&cfg.AppID, "app-id", "", "app UUID on omaha server (not pod UUID)")
	fs.StringVar(&cfg.Endpoint, "endpoint", "", "omaha server URL")
	fs.StringVar(&cfg.Group, "group", "", " omaha user configured tags, e.g. 'alpha'")
	fs.StringVar(&cfg.UnitDir, "unit-dir", "/run/systemd/system", "systemd unit path for writing payloads")
	fs.StringVar(&cfg.UpdateMethod, "update-method", "restart", "'restart' is the only valid method for now")
	fs.BoolVar(&cfg.InsecureSkipVerify, "insecure-skip-verify", false, "skip signature verification of payload")
	fs.Var(&cfg.PubKeys, "keys", "base64 encoded ed25519 public key, multiple flags will specify multiple keys ")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("flag parsing failed: %v", err)
	}
	if fs.NArg() != 0 {
		log.Fatalf("unknown arguments: %v", fs.Args())
	}

	getFlagsFromEnv(cliName, fs)

	if err := cfg.Valid(); err != nil {
		log.Fatalf("%v", err)
	}

	err := strudel.RunTryUpdate(cfg)
	if err != nil {
		log.Fatalf("failed to update: %v", err)
	}

	return
}

// getFlagsFromEnv parses all registered flags in the given flagset,
// and if they are not already set it attempts to set their values from
// environment variables. Environment variables take the name of the flag but
// are UPPERCASE, have the given prefix, and any dashes are replaced by
// underscores - for example: some-flag => PREFIX_SOME_FLAG
func getFlagsFromEnv(prefix string, fs *flag.FlagSet) {
	alreadySet := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		alreadySet[f.Name] = true
	})
	fs.VisitAll(func(f *flag.Flag) {
		if !alreadySet[f.Name] {
			key := strings.ToUpper(prefix + "_" + strings.Replace(f.Name, "-", "_", -1))
			val := os.Getenv(key)
			if val != "" {
				fs.Set(f.Name, val)
			}
		}

	})
}
