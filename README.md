# dht-store

[![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Package dhtstore provides a data store over dht.


# TOC
- [Install](#install)
- [Usage](#usage)
- [API example](#api-example)
  - [> demo/main.go](#-demomaingo)

# Install

```sh
go get github.com/mh-cbon/dht-store
```

# Usage

```sh
cd demo
# term1
go run main.go -add tomate -val 66 -kname test -salt abc
# term2
go run main.go -port 9091 -put tomate -val 666 -kname test -salt abc -seq 1
# term 3
go run main.go -port 9093 -get tomate -kname test -salt abc -seq 1
#kill term 2
go run main.go -port 9093 -get tomate -kname test -salt abc

# don t forget the data timeout is set to 10 sec!
```

# API example

#### > demo/main.go
```go
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/anacrolix/dht"
	"github.com/anacrolix/dht/krpc"
	dhtstore "github.com/mh-cbon/dht-store"
	"golang.org/x/crypto/ed25519"
)

func main() {

	port := flag.Int("port", 9090, "")
	put := flag.String("put", "", "")
	add := flag.String("add", "", "")
	val := flag.String("val", "", "")
	get := flag.String("get", "", "")
	kname := flag.String("kname", "", "")
	salt := flag.String("salt", "", "")
	seq := flag.Int("seq", -1, "")
	cas := flag.Int("cas", 0, "")
	flag.Parse()

	bootstrap := []string{"localhost:9090", "localhost:9091", "localhost:9092", "localhost:9093"}
	// if *port == 9093 {
	// 	bootstrap = []string{"localhost:9092"}
	// }
	addr := fmt.Sprintf("localhost:%v", *port)

	table, netErr := getDhtTable(addr, bootstrap, 10*time.Second)
	if netErr != nil {
		log.Panic(netErr)
	}
	keyer, keyErr := getKeyer(*kname, *salt)
	if keyErr != nil {
		log.Panic(keyErr)
	}
	store := dhtstore.NewStorerSync(keyer)
	connectedStore := dhtstore.NewConnectedStore(table, keyer, store)

	keepAlive := connectedStore.KeepAlive(4 * time.Second)
	go keepAlive.Start()

	<-time.After(1 * time.Second)

	if *add != "" {
		key, err := connectedStore.Add(*add, *val, 8)
		if err != nil {
			log.Panic(err)
		}
		fmt.Printf("%#v\n", key)
	}

	if *put != "" {
		key, err := connectedStore.Put(*put, *val, *seq, *cas, 8)
		if err != nil {
			log.Panic(err)
		}
		fmt.Printf("%#v\n", key)
	}

	if *get != "" {
		value, err := connectedStore.Get(*get, *seq, 8)
		if err != nil {
			log.Panic(err)
		}
		fmt.Printf("%#v\n", value)
		os.Exit(0)
	}
	<-make(chan bool)
}

func getDhtTable(addr string, bootstrap []string, refresh time.Duration) (*dht.Server, error) {
	conf := &dht.ServerConfig{
		Addr:               addr,
		NoDefaultBootstrap: !true,
		NoSecurity:         true,
		BootstrapNodes:     bootstrap,
		StoreTimeout:       refresh,
		OnQuery: func(query *krpc.Msg, source net.Addr) (propagate bool) {
			log.Printf("query %v %#v\n", source, query.Q)
			return true // true or false ? unclear doc.
		},
		OnResponsePeer: func(query krpc.Msg, source dht.Addr) {
			r := ""
			if query.R != nil {
				r = fmt.Sprintf("ret.V=%#v ret.Seq=%#v ret.K=%#v ret.Sign=%#v",
					query.R.V, query.R.Seq, query.R.K, query.R.Sign)
			}
			log.Printf("response %v err=%v %v\n", source, query.Error(), r)
		},
	}
	return dht.NewServer(conf)
}

func getKeyer(name, salt string) (dhtstore.Keyer, error) {
	var ret dhtstore.Keyer
	ret = dhtstore.ImmutableKeyer{}
	if name != "" {
		identity, err := getIdentity(name, salt)
		if err != nil {
			return nil, err
		}
		ret = dhtstore.NewMutableKeyer(identity)
	}
	return ret, nil
}

func getIdentity(name, salt string) (dhtstore.Identity, error) {
	var ret dhtstore.Identity
	file := name + salt + ".key"
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(file)
		return dhtstore.NewIdentity(ed25519.PrivateKey(b), name, salt), err
	}
	_, pvk, err := ed25519.GenerateKey(nil)
	if err != nil {
		return ret, err
	}
	return dhtstore.NewIdentity(pvk, name, salt), ioutil.WriteFile(file, pvk, os.ModePerm)
}
```
