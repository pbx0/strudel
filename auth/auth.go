package auth

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/agl/ed25519"
)

type (
	Sig     [ed25519.SignatureSize]byte
	PubKey  [ed25519.PublicKeySize]byte
	PubKeys []PubKey // implements flag.Value
)

// DecodeKey returns a PubKey from its base64 representation
func DecodeKey(s string) (PubKey, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return PubKey{}, err
	}
	if len(data) != ed25519.PublicKeySize {
		return PubKey{}, fmt.Errorf("keys are 32 bytes ed25519 public keys, recieved: %v", len(data))
	}

	// slice -> array
	var key PubKey
	copy(key[:], data)

	return key, nil
}

// DecodeSig returns a Sig from its base64 representation
func DecodeSig(s string) (Sig, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return Sig{}, err
	}
	if len(data) != ed25519.SignatureSize {
		return Sig{}, fmt.Errorf("keys are 32 bytes ed25519 public keys, recieved: %v", len(data))
	}

	// slice -> array
	var key Sig
	copy(key[:], data)

	return key, nil
}

func (key PubKey) Verify(msg []byte, sig Sig) bool {
	var (
		k [ed25519.PublicKeySize]byte = key
		s [ed25519.SignatureSize]byte = sig
	)
	return ed25519.Verify(&k, msg, &s)
}

func (keys PubKeys) String() string {
	var out []string
	for _, key := range keys {
		s := base64.StdEncoding.EncodeToString(key[:])
		out = append(out, s)
	}
	return fmt.Sprintf("%v", out)
}

// Set will append additional keys for each flag set. Comma separated flags
// without spaces will also be parsed correctly.
func (keys *PubKeys) Set(value string) error {
	values := strings.Split(value, ",")

	for _, s := range values {
		key, err := DecodeKey(s)
		if err != nil {
			return err
		}
		*keys = append(*keys, key)
	}

	return nil
}
