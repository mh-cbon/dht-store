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
