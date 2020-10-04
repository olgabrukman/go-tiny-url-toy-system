package main

import (
	"log"

	"go-tiny-url-toy-system/api/httpendpoint"
	"go-tiny-url-toy-system/app/cache"
)

func main() {
	myCache := cache.Cache{}
	err := myCache.Init(nil)
	if err != nil {
		log.Fatal("Failed to init cache, ", err)
	}

	server := httpendpoint.NewServer()
	server.Start(myCache.Collection)
	//hash, err := myCache.GetEncoding("ynet.com")
	//println( hash == "eW5ldC5jb20=")
	//
	//url, err := myCache.GetURL("eW5ldC5jb20=")
	//println( url == "ynet.com")
	select {}
}
