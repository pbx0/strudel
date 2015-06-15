package strudel

import (
	"bytes"
	"encoding/base64"
	"testing"
	"unicode"

	"github.com/agl/ed25519"
	"github.com/coreos/strudel/auth"
)

type zeroReader struct{}

func (zeroReader) Read(buf []byte) (int, error) {
	for i := range buf {
		buf[i] = 0
	}
	return len(buf), nil
}

var (
	zero               zeroReader
	public, private, _ = ed25519.GenerateKey(zero)
	pubString          = base64.StdEncoding.EncodeToString(public[:])
	badSigner          = base64.StdEncoding.EncodeToString(make([]byte, ed25519.PublicKeySize))
)

func TestPayloadVerify(t *testing.T) {
	var payloadTests = []struct {
		input         []byte // JSON with servicefile and signer fields, sig is added automatically
		expectSuccess bool
	}{
		{[]byte(`{"servicefile":"test", "signer":""}`), false},                  // empty signer
		{[]byte(`{"servicefile":"test", "signer":"` + badSigner + `"}`), false}, // signer doesn't match added sig
		{[]byte(`{"servicefile":"", "signer":"` + pubString + `"}`), true},      // empty service file
		{[]byte(`{"signer":"` + pubString + `"}`), true},                        // service file missing
		{[]byte(`{"servicefile":"test", "signer":"` + pubString + `"}`), true},
		{[]byte(`{"servicefile":"test","foo":"bar","signer":"` + pubString + `"}`), true}, // exta fields ok as long as sig is last
	}
	for _, tt := range payloadTests {
		// add signature
		input := addSig(tt.input)

		// parse input
		p, err := NewPayload(input)
		if err != nil {
			if tt.expectSuccess == false {
				continue
			} else {
				t.Fatalf("err: %v from NewPayload using test input: %s", err, input)
			}
		}
		// verify input using signer
		givenSigner, _ := auth.DecodeKey(p.Signer)
		actual := p.VerifySig(auth.PubKeys{givenSigner})
		if actual != tt.expectSuccess {
			t.Fatalf("expected %v from VerifySig using test input: %s", tt.expectSuccess, input)
		}
	}

}

func addSig(b []byte) []byte {

	// remove trailing whitespace
	b = bytes.TrimRightFunc(b, unicode.IsSpace)
	// and last '}'
	b = b[:len(b)-1]

	sig := ed25519.Sign(private, b)
	sigString := base64.StdEncoding.EncodeToString(sig[:])

	// append signature
	suffix := []byte(`,"sig":"` + sigString + `"}` + "\n")
	b = append(b, suffix...)

	return b
}
