package main

import (
	"log"

	"userms/internal/config"
	"userms/internal/server"
)

func main() {
	cfg := config.Load()
	srv := server.New(cfg)
	
	log.Fatal(srv.Start())
}