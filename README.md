# dht-store

[![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Package dhtstore provides a data store over dht.


# TOC
- [Install](#install)
- [Usage](#usage)

# Install

```sh
go get github.com/mh-cbon/dht-store
```

# Usage

```sh
cd demo
# term1: announce tomate=66
go run main.go -add tomate -val 66 -kname test -salt abc
# term2: uodate tomate=666
go run main.go -port 9091 -put tomate -val 666 -kname test -salt abc -seq 1
# print 666
# term 3:  get tomate value
go run main.go -port 9093 -get tomate -kname test -salt abc -seq 1
#kill term 2, then  get tomate value again
go run main.go -port 9093 -get tomate -kname test -salt abc
# print 66

# don t forget the data timeout is set to 10 sec!
```
