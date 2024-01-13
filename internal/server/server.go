package server

import (
	"context"
	"fmt"
	"github.com/A-ndrey/gistKV/internal/storage"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	nodePath = "/v1/node/"
	listPath = "/v1/list"
)

func Start(s *storage.Storage) (stopFunc func(duration time.Duration)) {
	http.HandleFunc(nodePath, nodeHandler(s))
	http.HandleFunc(listPath, listHandler(s))

	server := &http.Server{
		Addr: ":8080",
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return func(timeout time.Duration) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown: ", err)
		}
	}
}

func nodeHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimLeft(r.URL.Path, nodePath)

		_, force := r.URL.Query()["force"]

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		value := string(body)

		switch r.Method {
		case "GET":
			val, err := s.Read(key)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = fmt.Fprint(w, val)
			if err != nil {
				log.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
		case "POST":
			err = s.Create(key, value)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "PUT":
			err = s.Update(key, value)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "DELETE":
			err = s.Delete(key, force)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func listHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := r.URL.Query().Get("format")

		list, err := s.List(format)
		if err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		_, err = fmt.Fprint(w, list)
		if err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}
