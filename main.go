package main

import (
	"gistKV/gist"
	"gistKV/server"
	"gistKV/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")

	client, err := gist.NewClient(token)
	if err != nil {
		log.Fatalln(err)
	}

	s := storage.New(client)

	stopFunc := server.Start(s)
	defer stopFunc(5 * time.Second)

	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
