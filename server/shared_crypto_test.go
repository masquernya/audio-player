package server

import (
	"log"
	"testing"
)

func Test_createKeyPair(t *testing.T) {
	s := newShared()
	key, cert := s.createKeyPair("client")
	if len(key) == 0 {
		t.Error("key is empty")
	}
	if len(cert) == 0 {
		t.Error("cert is empty")
	}

	t.Log("cert:", cert)
	t.Log("key:", key)
}

func TestShared_GetKeyPair(t *testing.T) {
	s := newShared()
	cert := s.GetKeyPair(KeyPairModeClient)
	if cert == nil {
		t.Error("cert is nil")
	}
	log.Println("cert:", cert)
}
