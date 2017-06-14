package dhtstore

import (
	"log"
	"time"
)

type KeepAlive struct {
	store   *ConnectedStore
	keyer   Keyer
	refresh time.Duration
	done    chan bool
}

func KeepStoreAlive(store *ConnectedStore, keyer Keyer, refresh time.Duration) *KeepAlive {
	ret := &KeepAlive{
		store:   store,
		keyer:   keyer,
		refresh: refresh,
		done:    make(chan bool),
	}
	return ret
}
func (s *KeepAlive) Start() {
	for {
		select {
		case <-time.After(s.refresh):
			s.keepAlive()
		case <-s.done:
			return
		}
	}
}
func (s *KeepAlive) Stop() {
	s.done <- true
}
func (s *KeepAlive) keepAlive() {
	data := s.store.Map()
	if len(data) == 0 {
		log.Println("empty data")
	}
	for key, value := range data {
		// key, err := s.keyer.Key(name, value)
		// if err != nil {
		// 	log.Println(err)
		// 	continue
		// }
		// log.Println("name ", name)
		// log.Println("value ", value)
		// log.Println("key ", key)
		if !s.store.ClearStat(key) {
			log.Println("not cleared ", key)
			continue
		}
		stat := s.store.Stat(key)
		go func(name, value string) {
			_, putErr := s.store.Put(name, value, stat.LastSeq, stat.LastCas)
			if putErr != nil {
				log.Println(putErr)
			}
		}(stat.Name, value)
	}
}
