package main

import (
	"log"

	"github.com/imafaz/B9CA/controller"
)

func main() {
	cn := controller.NewBindManager(`/etc/bind/zones`, `/etc/bind/named.conf.local`)
	err := cn.AddDomain(`afaz.me`, "ns581.servercap.com", "ns591.servercap.com")
	if err != nil {
		log.Fatal(err)
	}
	err = cn.AddRecord(`afaz.me`, controller.A, "www", "8.8.8.8", 8600)
	if err != nil {
		log.Fatal(err)
	}
}
