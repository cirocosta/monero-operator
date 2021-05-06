// thanks https://github.com/rdkr/oniongen-go
//
package tor

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base32"
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"
)

const (
	FilenameHostname  = "hostname"
	FilenamePublicKey = "hs_ed25519_secret_key"
	FilenameSecretKey = "hs_ed25519_public_key"
)

type Credentials map[string][]byte

func GenerateCredentials() (Credentials, error) {
	publicKey, secretKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	onionAddress := encodePublicKey(publicKey)

	return Credentials{
		FilenameSecretKey: append([]byte("== ed25519v1-secret: type0 ==\x00\x00\x00"), secretKey[:]...),
		FilenamePublicKey: append([]byte("== ed25519v1-public: type0 ==\x00\x00\x00"), publicKey...),
		FilenameHostname:  []byte(onionAddress + ".onion"),
	}, nil
}

func encodePublicKey(publicKey ed25519.PublicKey) string {
	var (
		checksumBytes     bytes.Buffer
		onionAddressBytes bytes.Buffer
	)

	checksumBytes.Write([]byte(".onion checksum"))
	checksumBytes.Write([]byte(publicKey))
	checksumBytes.Write([]byte{0x03})
	checksum := sha3.Sum256(checksumBytes.Bytes())

	onionAddressBytes.Write([]byte(publicKey))
	onionAddressBytes.Write(checksum[:2])
	onionAddressBytes.Write([]byte{0x03})
	onionAddress := base32.StdEncoding.EncodeToString(onionAddressBytes.Bytes())

	return strings.ToLower(onionAddress)
}
