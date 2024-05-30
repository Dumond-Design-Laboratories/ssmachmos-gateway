package model

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func VerifySignature(data []byte, signature []byte, publicKey *rsa.PublicKey) bool {
	hash := sha256.Sum256(data)
	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	return err == nil
}

func ParsePublicKey(value []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(value)
	if block == nil {
		return nil, errors.New("failed to parse public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to parse public key")
	}

	return rsaPub, nil
}

func UuidToBytes(uuid [4]uint32) [16]byte {
	return [16]byte{
		byte(uuid[0] >> 24),
		byte(uuid[0] >> 16),
		byte(uuid[0] >> 8),
		byte(uuid[0]),
		byte(uuid[1] >> 24),
		byte(uuid[1] >> 16),
		byte(uuid[1] >> 8),
		byte(uuid[1]),
		byte(uuid[2] >> 24),
		byte(uuid[2] >> 16),
		byte(uuid[2] >> 8),
		byte(uuid[2]),
		byte(uuid[3] >> 24),
		byte(uuid[3] >> 16),
		byte(uuid[3] >> 8),
		byte(uuid[3]),
	}
}

func BytesToUuid(value [16]byte) [4]uint32 {
	return [4]uint32{
		uint32(value[0])<<24 | uint32(value[1])<<16 | uint32(value[2])<<8 | uint32(value[3]),
		uint32(value[4])<<24 | uint32(value[5])<<16 | uint32(value[6])<<8 | uint32(value[7]),
		uint32(value[8])<<24 | uint32(value[9])<<16 | uint32(value[10])<<8 | uint32(value[11]),
		uint32(value[12])<<24 | uint32(value[13])<<16 | uint32(value[14])<<8 | uint32(value[15]),
	}
}

func UuidToString(uuid [4]uint32) string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x", uuid[3], uuid[2]>>16, uuid[2]&0xffff, uuid[1]>>16, uuid[1]&0xffff, uuid[0])
}

func GenerateUUID() ([4]uint32, error) {
	u, err := uuid.New().MarshalBinary()
	if len(u) != 16 || err != nil {
		return [4]uint32{}, errors.New("failed to generate UUID")
	}
	return BytesToUuid([16]byte(u)), nil
}
