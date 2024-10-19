package server

import (
	crypto_rand "crypto/rand"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/nacl/secretbox"
	"log"
	"os"
	"path"
)

func (s *Shared) GetEncryptionKey() []byte {
	encryptionKeyPath := path.Join(path.Dir(os.Args[0]), "encryption-key")
	hexBits, err := os.ReadFile(encryptionKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("encryption key file does not exist, creating it")
			bits := make([]byte, 32)
			_, err = crypto_rand.Read(bits)
			if err != nil {
				log.Fatal("error generating encryption key:", err)
			}

			hexBits = []byte(hex.EncodeToString(bits))
			err = os.WriteFile(encryptionKeyPath, hexBits, 0600)
			if err != nil {
				log.Fatal("error writing encryption key file:", err)
			}
		} else {
			log.Fatal("error reading encryption key file:", err)
		}
	}
	// encoded as hex
	bits, err := hex.DecodeString(string(hexBits))
	if err != nil {
		log.Fatal("error decoding encryption key file:", err)
	}
	return bits
}

func (s *Shared) Encrypt(b []byte) []byte {
	nonce := [24]byte{}
	_, err := crypto_rand.Read(nonce[:])
	if err != nil {
		log.Fatal("error generating nonce:", err)
	}

	key := [32]byte{}
	copy(key[:], s.GetEncryptionKey())
	out := secretbox.Seal(nil, b, &nonce, &key)
	out = append(nonce[:], out...)
	return out
}

func (s *Shared) Decrypt(b []byte) ([]byte, error) {
	if len(b) < 24 {
		return nil, errors.New("message is too short")
	}

	nonce := [24]byte{}
	copy(nonce[:], b[:24])

	key := [32]byte{}
	copy(key[:], s.GetEncryptionKey())

	out, ok := secretbox.Open(nil, b[24:], &nonce, &key)
	if !ok {
		return nil, errors.New("could not decrypt message")
	}
	return out, nil
}
