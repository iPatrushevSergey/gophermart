package main

import (
	"log"

	"gophermart/cmd/gophermart/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatalf("gophermart: %v", err)
	}
}
