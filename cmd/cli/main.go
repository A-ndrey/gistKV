package main

import (
	"fmt"
	"github.com/A-ndrey/gistKV/internal/gist"
	"github.com/A-ndrey/gistKV/internal/storage"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("specify one of the commands: create, read, update, delete, list")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatalln("specify GITHUB_TOKEN env")
	}

	gistClient, err := gist.NewClient(token)
	if err != nil {
		log.Fatalln(err)
	}

	s := storage.New(gistClient)

	cmd := os.Args[1]
	switch strings.ToLower(strings.TrimSpace(cmd)) {
	case "create":
		if len(os.Args) != 4 {
			log.Fatalln("invalid number of parameters")
		}
		err := s.Create(os.Args[2], os.Args[3])
		if err != nil {
			log.Fatalln(err)
		}
	case "read":
		if len(os.Args) != 3 {
			log.Fatalln("invalid number of parameters")
		}
		res, err := s.Read(os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(res)
	case "update":
		if len(os.Args) != 4 {
			log.Fatalln("invalid number of parameters")
		}
		err := s.Update(os.Args[2], os.Args[3])
		if err != nil {
			log.Fatalln(err)
		}
	case "delete":
		if len(os.Args) != 3 {
			log.Fatalln("invalid number of parameters")
		}
		err := s.Delete(os.Args[2], true)
		if err != nil {
			log.Fatalln(err)
		}
	case "list":
		if len(os.Args) != 2 {
			log.Fatalln("invalid number of parameters")
		}
		res, err := s.List(storage.JsonFormat)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(res)
	default:
		log.Fatalln("unknown command")
	}
}
