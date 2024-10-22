package server

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"strconv"
	"time"
)

type KeyPairMode int

const (
	KeyPairModeServer KeyPairMode = 1
	KeyPairModeClient             = 2

	commonNameSuffix = "_linux_audio_player_1"
)

// createKeyPair creates a key for modeStr and returns (privateKey, certificate)
func (s *Shared) createKeyPair(modeStr string) (string, string) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal("error generating key pair:", err)
	}

	base := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: modeStr + commonNameSuffix,
		},
		SerialNumber: big.NewInt(time.Now().UnixMilli()),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
	}
	if modeStr == "server" {
		base.IPAddresses = append(base.IPAddresses, net.ParseIP("127.0.0.1"))
	}
	generatedCert, err := x509.CreateCertificate(nil, base, base, pubKey, privKey)
	if err != nil {
		log.Fatal("error creating "+modeStr+" cert:", err)
	}

	key, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		log.Fatal("error marshaling "+modeStr+" key:", err)
	}

	certStr := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: generatedCert,
	})

	keyStr := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: key,
	})

	//certStr := base64.StdEncoding.EncodeToString(generatedCert)
	//keyStr := base64.StdEncoding.EncodeToString(key)

	return string(keyStr), string(certStr)
}

func (s *Shared) GetKeyPair(mode KeyPairMode) tls.Certificate {
	s.certMux.Lock()
	defer s.certMux.Unlock()

	modeStr := "client"
	if mode == KeyPairModeServer {
		modeStr = "server"
	}

	expectedCertPath := path.Join(path.Dir(os.Args[0]), modeStr+".crt")
	expectedKeyPath := path.Join(path.Dir(os.Args[0]), modeStr+".key")

	doCreate := false

	_, err := os.Stat(expectedCertPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal("error reading "+modeStr+" cert:", err)
		}

		// create a new pair
		doCreate = true
	}

	_, err = os.Stat(expectedKeyPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal("error reading "+modeStr+" key:", err)
		}

		// create a new pair
		doCreate = true
	}

	if !doCreate {
		cert, err := tls.LoadX509KeyPair(expectedCertPath, expectedKeyPath)
		if err != nil {
			log.Fatal("error loading keys for "+modeStr, err)
		}

		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			log.Fatal("error parsing cert for "+modeStr, err)
		}

		if cert.Leaf.NotAfter.Before(time.Now().AddDate(0, 0, 30)) {
			log.Println("request regen for " + modeStr + " cert")
			doCreate = true
		}
	}

	if doCreate {
		keyStr, certStr := s.createKeyPair(modeStr)

		if err := os.WriteFile(expectedCertPath, []byte(certStr), 0600); err != nil {
			log.Fatal("error writing "+modeStr+" cert:", err)
		}

		if err := os.WriteFile(expectedKeyPath, []byte(keyStr), 0600); err != nil {
			log.Fatal("error writing "+modeStr+" key:", err)
		}
	}

	cert, err := tls.LoadX509KeyPair(expectedCertPath, expectedKeyPath)
	if err != nil {
		log.Fatal("error loading keys for "+modeStr, err)
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		log.Fatal("error parsing cert for "+modeStr, err)
	}

	return cert
}

func (s *Shared) GetCertificates() (tls.Certificate, tls.Certificate) {
	return s.GetKeyPair(KeyPairModeServer), s.GetKeyPair(KeyPairModeClient)
}

func (s *Shared) verifyCert(accessor func() tls.Certificate) func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	expected := accessor()
	expectedHash := sha256.Sum256(expected.Certificate[0])

	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		if len(rawCerts) != 1 {
			log.Println("got request with other than 1 cert: " + strconv.Itoa(len(rawCerts)))
			return errors.New("wrong cert count")
		}

		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			log.Println("error parsing cert", err)
			return err
		}

		// should not happen since we add it in GetKeyPair
		if expected.Leaf == nil {
			log.Fatal("expected cert has no leaf")
		}

		// check if expected cert has expired.
		if expected.Leaf.NotAfter.Before(time.Now()) {
			log.Println("expected cert has expired. requesting a new one.")
			expected = accessor()
			expectedHash = sha256.Sum256(expected.Certificate[0])
		}

		sha256Hash := sha256.Sum256(cert.Raw)
		if sha256Hash == expectedHash {
			return nil
		}
		log.Println("got cert with different hash:", sha256Hash, " vs ", expectedHash)

		return errors.New("invalid cert")
	}
}
