package main

import (
	"github.ocm/A-ndrey/gistKV/internal/gist"
	"github.ocm/A-ndrey/gistKV/internal/server"
	"github.ocm/A-ndrey/gistKV/internal/storage"
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
