package strudel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/coreos/strudel/auth"
)

// Omaha payload is json and signed using the camlistore spec:
// https://github.com/camlistore/camlistore/tree/master/doc/json-signing/example
type Payload struct {
	ServiceFile string `json:"servicefile"` // base64 encoded service file
	Signer      string `json:"signer"`      // base64 ed25519 pubkey
	Sig         string `json:"sig"`         // base64 ed25519 sig

	signedBytes []byte `json:"-"` // actual signed data
}

// NewPayload takes the byte payload of an omaha update and parses it according
// to the camlistore json-signing spec. It does not verify the signature.
func NewPayload(jsonbytes []byte) (*Payload, error) {
	const sep = `,"sig":"`

	//bytes all
	ba := jsonbytes

	// find last index in b of the 8 byte substring: `,"sig":"`
	idx := bytes.LastIndex(ba, []byte(sep))
	if idx == -1 {
		return nil, fmt.Errorf("no 8 byte separator found in json")
	}

	// TODO: rewrite camli sig spec to embody our name changes(e.g. camliSig ->
	// sig) and change these ba,bp,bpj names to something less confusing. For
	// now, the names might as well make sense while reading  camlistore spec.

	// bytes payload
	bp := ba[:idx]

	// bytes payload json
	bpj := ba[:idx+1]
	bpj[idx] = '}'

	// bytes sig. Must allocate new underlying memory for sig because otherwise
	// the "}" set in bs will overwrite the "{" we would set in bpj.
	bs := []byte("{")
	bs = append(bs, ba[idx+1:]...)

	// parse bpj as a Payload
	var p Payload
	if err := json.Unmarshal(bpj, &p); err != nil {
		return nil, fmt.Errorf("invalid JSON in payload: %v", err)
	}
	if p.Signer == "" {
		return nil, fmt.Errorf("empty or missing 'signer' field in JSON payload")
	}

	// parse bs as JSON and ensure that only the key "sig" exists
	var sigMap map[string]interface{}
	if err := json.Unmarshal(bs, &sigMap); err != nil {
		return nil, fmt.Errorf("invalid JSON in signature: %v", err)
	}

	if len(sigMap) != 1 {
		return nil, fmt.Errorf("JSON signature must have 1 key")
	}

	sig, ok := sigMap["sig"]
	if !ok {
		return nil, fmt.Errorf("missing 'sig' in JSON signature")
	}

	p.Sig, ok = sig.(string)
	if !ok {
		return nil, fmt.Errorf("'sig' is not a string")
	}

	p.signedBytes = bp
	return &p, nil

}

// VerifySig matches the Payload.Signer field with one of the provided trusted public
// keys and cryptographically verifies the payload signature using that public
// key.
func (p *Payload) VerifySig(keys auth.PubKeys) bool {
	signer, err := auth.DecodeKey(p.Signer)
	if err != nil {
		log.Print(err)
		return false
	}
	sig, err := auth.DecodeSig(p.Sig)
	if err != nil {
		log.Print(err)
		return false
	}

	// verify signer is a trusted key
	var signerTrusted bool
	for _, key := range keys {
		if key == signer {
			signerTrusted = true
			break
		}
	}
	if !signerTrusted {
		return false
	}

	return signer.Verify(p.signedBytes, sig)

}

// overwrite or create new service file
func (p *Payload) OverwriteServiceFile(path string) error {
	b, err := base64.StdEncoding.DecodeString(p.ServiceFile)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, b, os.FileMode(0644))
}
