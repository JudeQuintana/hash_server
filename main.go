package main

import (
	"github.com/JudeQuintana/hash_server/hasher"
	"log"
)

func main() {

	address, port := "localhost", "8080"
	hashServer, shutdown := hasher.NewHashServer(address, port)

	// Run HashServer in the background
	go func() {
		if err := hashServer.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// block main thread until channel is closed by /shutdown endpoint
	<-shutdown
}
