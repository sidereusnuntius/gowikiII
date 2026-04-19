package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/sidereusnuntius/gowiki/internal/consts"
)

func generateRSAKeyPair() (pub []byte, priv []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, consts.RsaKeySize)
	if err != nil {
		err = fmt.Errorf("failed to generate RSA key: %w", err)
		return
	}

	// Encode private key.
	derPriv, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		err = fmt.Errorf("failed to marshal RSA private key: %w", err)
		return
	}

	priv = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: derPriv,
	})

	// Encode public key.
	derPub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		err = fmt.Errorf("failed to marshal RSA public key: %w", err)
		return
	}

	pub = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPub,
	})

	return
}
