package keystore

import (
	"github.com/pborman/uuid"
	"log"

	"theta/common"
	"theta/crypto"
)

type Key struct {
	Id         uuid.UUID
	Address    common.Address
	PrivateKey *crypto.PrivateKey
}

func NewKey(privKey *crypto.PrivateKey) *Key {
	Id := uuid.NewRandom()
	return &Key{
		Id:         Id,
		Address:    privKey.PublicKey().Address(),
		PrivateKey: privKey,
	}
}

func (key *Key) Sign(data common.Bytes) (*crypto.Signature, error) {
	log.Println(key.PrivateKey.ToBytes(), " wallet/softwallet/keystore/key.go:27")
	sig, err := key.PrivateKey.Sign(data)
	return sig, err
}
