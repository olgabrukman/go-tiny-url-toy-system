package httpendpoint

import (
	"context"
	"log"
	"net/http"
	"time"

	"go-tiny-url-toy-system/app/cache"
	"go-tiny-url-toy-system/db/dbconnection"
)

const Addr = "localhost:8090"

type Server struct {
	httpServer *http.Server
	cache      *cache.Cache
}

func NewServer() *Server {
	s := Server{}
	s.cache = &cache.Cache{}
	return &s
}

func (s *Server) Start(collection *dbconnection.MyCollection) {
	if err := s.cache.Init(collection); err != nil {
		log.Fatalf("Failed to start cache, aborting; error: %v", err)
	}
	s.httpServer = &http.Server{Addr: Addr, Handler: s.routes()}
	//linter:errcheck
	go s.httpServer.ListenAndServe() //Blocks, thus need go

	log.Printf("Started http server on addr %s", Addr)
}

func (s *Server) Shutdown() {
	if s.httpServer != nil {
		//nolint: govet
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			log.Fatal("Failed to shutdown http Server gracefully")
		} else {
			s.httpServer = nil
		}
	}
}

func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", s.HandleShorten())
	mux.HandleFunc("/redirect", s.HandleRedirect())
	return mux
}

func (s *Server) HandleShorten() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, ok := extractParamFromRequest(r, "url")
		if !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		log.Printf("Handle shorten for %s\n", url)
		encoding, err := s.cache.GetEncoding(url)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			log.Printf("Failed to retrieve/compute url encoding for %s\n", url)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(encoding))
		log.Println(err)
	}
}

func extractParamFromRequest(r *http.Request, paramName string) (string, bool) {
	keys, ok := r.URL.Query()[paramName]
	if !ok || len(keys[0]) < 1 {
		log.Printf("URL Param '%s' is missing\n", paramName)
		return "", false
	}
	key := keys[0]
	log.Printf("URL Param 'key' is %s\n ", key)
	return key, true
}

func (s *Server) HandleRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash, ok := extractParamFromRequest(r, "hash")
		log.Printf("Handle redirect for %s\n", hash)
		if !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, err := w.Write([]byte("Failed to compute url from hash"))
			log.Println(err)
			return
		}
		url, err := s.cache.GetURL(hash)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			if _, err := w.Write([]byte("Failed to retrieve url from cache due to connection error")); err != nil {
				log.Println(err)
			}
		} else {
			//http.Redirect(w, r, "http://"+url, http.StatusSeeOther)
			w.WriteHeader(http.StatusSeeOther)
			w.Write([]byte(url))
			log.Println("Redirected request to url:", url)
			return
		}
	}
}
