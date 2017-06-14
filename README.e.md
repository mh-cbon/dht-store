---
License: MIT
LicenseFile: LICENSE
LicenseColor: yellow
---
# {{.Name}}

{{template "license/shields" .}}

{{pkgdoc "store.go"}}

# {{toc 5}}

# Install

{{template "go/install" .}}

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
