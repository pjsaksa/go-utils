package csp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func NewNonce() Nonce {
	var data [randomBytes]byte
	rand.Read(data[:])

	return Nonce(base64.URLEncoding.EncodeToString(data[:]))
}

const randomBytes = 12 // --> 16 bytes in base64

type Nonce string

func (n Nonce) cspValue() string {
	return fmt.Sprintf("'nonce-%s'", string(n))
}

func (n Nonce) Empty() bool {
	return len(string(n)) == 0
}

func (n Nonce) String() string {
	return string(n)
}
