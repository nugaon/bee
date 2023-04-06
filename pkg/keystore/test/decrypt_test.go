package test

import (
	"fmt"
	"testing"

	"github.com/ethersphere/bee/pkg/keystore/file"
)

func TestTransactionStoredTransaction(t *testing.T) {
	s := file.New("whatever")
	keyFile := []byte("{\"address\":\"8e3cb0148c5f39577fb815dc8c37795e30f5dcfa\",\"crypto\":{\"cipher\":\"aes-128-ctr\",\"ciphertext\":\"c3c64c5b4371a59d87a4883b499a5622774511d4421298c0e8849acb64d1111e\",\"cipherparams\":{\"iv\":\"e31330130d3045f1b2614e819a4c06f6\"},\"kdf\":\"scrypt\",\"kdfparams\":{\"n\":32768,\"r\":8,\"p\":1,\"dklen\":32,\"salt\":\"25be64863839d78f078ff7c874412c682dc2fc290138ce749ac3bdc7ed267b5a\"},\"mac\":\"c8ecf497e8afec977de05b2c514194aa329eca701ca559520a0d4545591e3d2f\"},\"version\":3}")
	key, _ := s.DecryptKeyTest(keyFile, "password")
	fmt.Printf("key: %x\n", key.D)
}
