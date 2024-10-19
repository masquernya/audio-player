package server

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	s := newShared()
	b := []byte("hello")

	encrypted := s.Encrypt(b)
	decrypted, err := s.Decrypt(encrypted)
	if err != nil {
		t.Fatal("error decrypting:", err)
	}
	if string(decrypted) != "hello" {
		t.Fatal("decrypted message is not the same as original message")
	}
}
