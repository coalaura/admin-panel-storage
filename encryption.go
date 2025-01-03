package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type SessionKey struct {
	key   []byte
	block cipher.Block
}

type EncryptionKey struct {
	public *rsa.PublicKey
}

func NewEncryptionKey(pemData []byte) (*EncryptionKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("invalid public key PEM format")
	}

	public, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &EncryptionKey{
		public: public,
	}, nil
}

func (k *EncryptionKey) Encrypt(data []byte) ([]byte, error) {
	hash := sha256.New()

	return rsa.EncryptOAEP(hash, rand.Reader, k.public, data, nil)
}

func NewSessionKey() (*SessionKey, error) {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &SessionKey{
		key:   key,
		block: block,
	}, nil
}

func (k *SessionKey) Encrypt(data []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(k.block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func (k *SessionKey) Decrypt(data []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(k.block)
	if err != nil {
		return nil, err
	}

	size := gcm.NonceSize()

	return gcm.Open(nil, data[:size], data[size:], nil)
}
