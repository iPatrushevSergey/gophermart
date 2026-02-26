package main

import (
	"flag"
	"log"

	"gophermart/cmd/gophermart/bootstrap"
)

func main() {
	dsn := flag.String("d", "", "database dsn")
	dir := flag.String("dir", "../migrations/gophermart", "path to migration files")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("migrate: -d is required")
	}

	if err := bootstrap.RunMigrations(*dsn, *dir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}
