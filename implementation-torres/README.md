# Golang basic project

## spec
- [Golang](https://golang.org/) - usign [GVM](https://github.com/moovweb/gvm)
- [dep](https://github.com/golang/dep)
- [Makefile](https://wiki.openoffice.org/wiki/MakeFile)

## service
- [service-gateway](./service-gateway)
- [service-front](./service-front)
- RDB

### start
- install & prepare
```bash
make prepare # install
make mycli # use mycli to connect MySQL
```
- start docker container
```bash
make compose
```