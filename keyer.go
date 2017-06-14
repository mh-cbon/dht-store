package dhtstore

import (
	"github.com/anacrolix/dht"
)

type Keyer interface {
	Key(name, value string) (string, error)
	PutMessage(name, value string) (dht.MessageEncoder, error)
	GetMessage(name string) (dht.MessageEncoderChecker, error)
}

type ImmutableKeyer struct{}

func (s ImmutableKeyer) Key(name, value string) (string, error) {
	return hashSha1(value), nil
}

func (s ImmutableKeyer) PutMessage(name, value string) (dht.MessageEncoder, error) {
	return dht.NewPutMessage(value), nil
}

func (s ImmutableKeyer) GetMessage(key string) (dht.MessageEncoderChecker, error) {
	return dht.NewGetMessage(key), nil
}

func NewMutableKeyer(i Identity) MutableKeyer {
	return MutableKeyer{identity: i}
}

type MutableKeyer struct {
	identity Identity
}

func (s MutableKeyer) Key(name, value string) (string, error) {
	return hashSha1(string(s.identity.Public()), s.identity.salt+name), nil
}

func (s MutableKeyer) PutMessage(name, value string) (dht.MessageEncoder, error) {
	ret := dht.NewMutablePutMessage(value)
	err := ret.PrivateKey(s.identity.pvk, s.identity.salt+name)
	return ret, err
}

func (s MutableKeyer) GetMessage(name string) (dht.MessageEncoderChecker, error) {
	return dht.NewMutableGetMessage(s.identity.Public(), s.identity.salt+name)
}
