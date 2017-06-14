// Package dhtstore provides a data store over dht.
package dhtstore

import (
	"errors"
	"time"
)

var NotFound = errors.New("not found")

type StoreValue struct {
	Value              string
	Name               string
	CreateDate         time.Time
	UpdateDate         time.Time
	WishReplicationCnt int
	ReplicationCount   int
	LastSeq            int
	LastCas            int
	Errors             []error
}

type Storer interface {
	Add(name, value string, copyCnt ...int) (key string, err error)
	Get(key string) (value string, err error)
	Remove(key string) error
	Keys() []string
	Values() []string
	Map() map[string]string
	Stats(keys ...string) []StoreValue
	Stat(key string) StoreValue
	AddErr(key string, err error) (added bool)
	UpdateReplicationCount(key string, count int) (newCount int)
	UpdateSeq(key string, seq int)
	UpdateCas(key string, cas int)
	ClearStat(key string) (cleared bool)
	Transact(...func(*Store))
}

func NewStore(keyer Keyer) *Store {
	return &Store{
		keyer: keyer,
		local: map[string]StoreValue{},
	}
}

type Store struct {
	keyer Keyer
	local map[string]StoreValue
}

func (s *Store) Transact(F ...func(*Store)) {
	for _, f := range F {
		f(s)
	}
}

func (s *Store) Add(name, value string, copyCnt ...int) (string, error) {
	key, err := s.keyer.Key(name, value)
	if err != nil {
		return key, err
	}
	if _, ok := s.local[key]; !ok {
		var c int
		if len(copyCnt) > 0 {
			c = copyCnt[0]
		}
		s.local[key] = StoreValue{
			Value:              value,
			Name:               name,
			WishReplicationCnt: c,
			CreateDate:         time.Now(),
			UpdateDate:         time.Now(),
		}
		return key, nil
	}
	return key, NotFound
}
func (s *Store) Get(key string) (string, error) {
	if _, ok := s.local[key]; ok {
		return s.local[key].Value, nil
	}
	return "", NotFound
}
func (s *Store) Remove(key string) error {
	if _, ok := s.local[key]; ok {
		delete(s.local, key)
		return nil
	}
	return NotFound
}
func (s *Store) Keys() []string {
	var ret []string
	for k := range s.local {
		ret = append(ret, k)
	}
	return ret
}
func (s *Store) Values() []string {
	var ret []string
	for _, v := range s.local {
		ret = append(ret, v.Value)
	}
	return ret
}
func (s *Store) Map() map[string]string {
	ret := map[string]string{}
	for k, v := range s.local {
		ret[k] = v.Value
	}
	return ret
}
func (s *Store) Stats(keys ...string) []StoreValue {
	ret := []StoreValue{}
	for key, v := range s.local {
		add := len(key) == 0
		if !add {
			for _, kk := range keys {
				if kk == key {
					add = true
					break
				}
			}
		}
		if add {
			ret = append(ret, v)
		}
	}
	return ret
}
func (s *Store) Stat(key string) StoreValue {
	ret := StoreValue{}
	if x, ok := s.local[key]; ok {
		ret = x
	}
	return ret
}
func (s *Store) ClearStat(key string) bool {
	if x, ok := s.local[key]; ok {
		x.Errors = x.Errors[:0]
		x.ReplicationCount = 0
		s.local[key] = x
		return true
	}
	return false
}
func (s *Store) UpdateReplicationCount(key string, count int) int {
	var ret int
	if x, ok := s.local[key]; ok {
		x.ReplicationCount += count
		if x.ReplicationCount < 0 {
			x.ReplicationCount = 0
		}
		s.local[key] = x
		ret = x.ReplicationCount
	}
	return ret
}
func (s *Store) UpdateSeq(key string, seq int) {
	if x, ok := s.local[key]; ok {
		x.LastSeq = seq
		s.local[key] = x
	}
}
func (s *Store) UpdateCas(key string, cas int) {
	if x, ok := s.local[key]; ok {
		x.LastCas = cas
		s.local[key] = x
	}
}
func (s *Store) AddErr(key string, err error) bool {
	if err == nil {
		return false
	}
	if x, ok := s.local[key]; ok {
		x.Errors = append(x.Errors, err)
		s.local[key] = x
		return true
	}
	return false
}
