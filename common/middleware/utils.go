package middleware

import (
	"errors"
	"github.com/gtank/cryptopasta"
)

func formKey(s string) (key [32]byte, err error) {
	if len(s) < 32 {
		return key, errors.New("String must be at least len 32")
	}

	sb := []byte(s)
	for i := 0; i < 32; i++ {
		key[i] = sb[i]
	}
	return
}

func decrypt(cypher []byte, key *[32]byte) ([]byte, error) {
	return cryptopasta.Decrypt(cypher, key)
}

func encrypt(data []byte, key *[32]byte) ([]byte, error) {
	return cryptopasta.Encrypt(data, key)
}
