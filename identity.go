package dhtstore

import "golang.org/x/crypto/ed25519"

type Identity struct {
	pvk  ed25519.PrivateKey
	name string
	salt string
}

func (i Identity) Public() ed25519.PublicKey {
	return i.pvk.Public().(ed25519.PublicKey)
}

func NewIdentity(pvk ed25519.PrivateKey, name, salt string) Identity {
	return Identity{
		pvk:  pvk,
		name: name,
		salt: salt,
	}
}
