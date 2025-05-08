package main

import (
	"context"
	"lockbox-webserver/db"
	"lockbox-webserver/web"
)

func main() {
	dbPool, err := db.NewPool(context.Background(), "postgresql://localhost:5432/lockbox")
	if err != nil {
		panic(err)
	}

	server, err := web.NewHTTPServer("http://localhost:8000", dbPool, []byte("abcdef"))
	if err != nil {
		panic(err)
	}

	server.Run(context.Background(), "127.0.0.1:8000")
}
