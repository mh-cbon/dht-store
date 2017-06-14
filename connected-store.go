package dhtstore

import (
	"strings"
	"time"

	"github.com/anacrolix/dht"
	"github.com/anacrolix/torrent/bencode"
)

func NewConnectedStore(table *dht.Server, keyer Keyer, store Storer) *ConnectedStore {
	return &ConnectedStore{
		table: table,
		keyer: keyer,
		store: store,
	}
}

type ConnectedStore struct {
	table *dht.Server
	keyer Keyer
	store Storer
}

func (s *ConnectedStore) KeepAlive(refresh time.Duration) *KeepAlive {
	return KeepStoreAlive(s, s.keyer, refresh)
}

func (s *ConnectedStore) Add(name, value string, copyCnt ...int) (string, error) {
	key, err := s.Put(name, value, -1, 0, copyCnt...)
	return key, err
}

func (s *ConnectedStore) Update(name, value string, copyCnt ...int) error {
	key, err := s.keyer.Key(name, value)
	if err != nil {
		return err
	}
	stat := s.store.Stat(key)
	_, err = s.Put(name, value, stat.LastSeq+1, stat.LastCas, copyCnt...)
	return err
}

type sequencer interface {
	Seq(seq int)
}
type caSwaper interface {
	Cas(cas int)
}

func (s *ConnectedStore) Put(name, value string, seq, cas int, copyCnt ...int) (string, error) {
	key, err := s.keyer.Key(name, value)
	if err != nil {
		return key, err
	}
	msg, errMsg := s.keyer.PutMessage(name, value)
	if errMsg != nil {
		return key, errMsg
	}
	if x, ok := msg.(sequencer); ok {
		x.Seq(seq)
	}
	if x, ok := msg.(caSwaper); ok {
		x.Cas(cas)
	}
	res, errPut := s.table.Put(msg, copyCnt...)
	if errPut != nil {
		return key, errPut
	}
	go func() {
		for {
			select {
			case putResponse, ok := <-res:
				if ok == false {
					return
				}
				s.store.Transact(func(store *Store) {
					if err := putResponse.Error(); err != nil {
						store.AddErr(key, err)
					} else {
						store.Add(name, value, copyCnt...)
						store.UpdateReplicationCount(key, 1)
						store.UpdateSeq(key, seq)
						store.UpdateCas(key, cas)
					}
				})
			}
		}
	}()
	return key, err
}
func (s *ConnectedStore) Get(key string, seq int, readCnt ...int) (string, error) {
	value, err := s.store.Get(key)
	if err != nil {
		value, err = s.Fetch(key, seq, readCnt...)
	}
	return value, err
}
func (s *ConnectedStore) Fetch(key string, seq int, readCnt ...int) (string, error) {
	msg, errMsg := s.keyer.GetMessage(key)
	if errMsg != nil {
		return key, errMsg
	}
	if x, ok := msg.(sequencer); ok {
		x.Seq(seq)
	}
	res, errGet := s.table.Get(msg, readCnt...)
	if errGet != nil {
		return key, errGet
	}
	cErr := make(chan error)
	cVal := make(chan string)
	go func() {
		errs := []error{}
		done := false
		doBreak := false
		for {
			select {
			case getResponse, ok := <-res:
				if ok == false {
					doBreak = true
				} else {
					if getResponse.Err != nil {
						errs = append(errs, getResponse.Err)
					} else {
						var val string
						decErr := bencode.NewDecoder(strings.NewReader(getResponse.Msg.V)).Decode(&val)
						if decErr != nil {
							errs = append(errs, decErr)
						} else if !done {
							cVal <- val
							done = true
							doBreak = true
							cErr <- nil
						}
					}
				}
			case <-time.After(10 * time.Second):
				doBreak = len(errs) > 0
			}
			if doBreak {
				break
			}
		}
		if !done {
			cVal <- ""
			if len(errs) > 0 {
				cErr <- errs[0]
			} else {
				cErr <- NotFound
			}
		}
	}()
	return <-cVal, <-cErr
}
func (s *ConnectedStore) Remove(key string) error {
	return s.store.Remove(key)
}
func (s *ConnectedStore) Keys() []string {
	return s.store.Keys()
}
func (s *ConnectedStore) Values() []string {
	return s.store.Values()
}
func (s *ConnectedStore) Map() map[string]string {
	return s.store.Map()
}
func (s *ConnectedStore) Stats() []StoreValue {
	return s.store.Stats()
}
func (s *ConnectedStore) ClearStat(key string) bool {
	return s.store.ClearStat(key)
}
func (s *ConnectedStore) Stat(key string) StoreValue {
	return s.store.Stat(key)
}
