package model

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
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

func UuidToBytes(uuid [4]uint32) []byte {
	result := []byte{}
	result = binary.LittleEndian.AppendUint32(result, uuid[0])
	result = binary.LittleEndian.AppendUint32(result, uuid[1])
	result = binary.LittleEndian.AppendUint32(result, uuid[2])
	result = binary.LittleEndian.AppendUint32(result, uuid[3])
	return result
}

func BytesToUuid(value [16]byte) [4]uint32 {
	return [4]uint32{
		binary.LittleEndian.Uint32(value[0:4]),
		binary.LittleEndian.Uint32(value[4:8]),
		binary.LittleEndian.Uint32(value[8:12]),
		binary.LittleEndian.Uint32(value[12:16]),
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
