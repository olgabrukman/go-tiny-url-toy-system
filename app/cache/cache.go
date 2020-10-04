package cache

import (
	"fmt"
	"log"
	"sync"
	"time"

	"go-tiny-url-toy-system/app/urlshortener"
	"go-tiny-url-toy-system/db/dbconnection"
)

const configFileName = "config"
const ImaxTTL = int64(time.Second * 5)
const tickDuration = time.Millisecond

type EncodingItem struct {
	encoding  string
	createdAt int64
}

type URLItem struct {
	url       string
	createdAt int64
}

type Cache struct {
	Collection       *dbconnection.MyCollection
	URLToEncodingMap map[string]*EncodingItem
	EncodingToURLMap map[string]*URLItem
	Mutex            sync.Mutex
}

func (cache *Cache) Init(collection *dbconnection.MyCollection) error {
	if collection == nil {
		collectionTmp, err := dbconnection.NewDBCollectionConnection(configFileName)
		if err != nil {
			return err
		}

		collection = collectionTmp
	}

	cache.Collection = collection
	cache.URLToEncodingMap = make(map[string]*EncodingItem)
	cache.EncodingToURLMap = make(map[string]*URLItem)
	go cache.clearOutdatedRecords()
	return nil
}

func (cache *Cache) clearOutdatedRecords() {
	log.Println("Cache clearance thread started")

	for now := range time.Tick(tickDuration) {
		nowTime := now.UnixNano()

		cache.Mutex.Lock()

		for url, item := range cache.URLToEncodingMap {
			if nowTime-item.createdAt > ImaxTTL {
				fmt.Println("clearing cache, removing url-encoding", url, item.encoding)
				delete(cache.URLToEncodingMap, url)
				delete(cache.EncodingToURLMap, item.encoding)
			}
		}
		cache.Mutex.Unlock()
	}
}

// Put url and its encoding to cache and DB if it was not stored yet
func (cache *Cache) putWithDBThroughput(url string) (string, error) {
	log.Println("writing data to db and cache")
	encoding := urlshortener.Encode(url)
	nowTime := time.Now().UnixNano()
	encodingItem := &EncodingItem{encoding, nowTime}
	urlItem := &URLItem{url, nowTime}

	_, err := cache.Collection.Insert(url, encoding)
	if err != nil {
		return "", err
	}

	cache.Mutex.Lock()
	_, ok := cache.URLToEncodingMap[url]
	if !ok {
		cache.URLToEncodingMap[url] = encodingItem
		cache.EncodingToURLMap[encoding] = urlItem
	} else {
		cache.URLToEncodingMap[url].createdAt = nowTime
		cache.EncodingToURLMap[encoding].createdAt = nowTime
	}
	cache.Mutex.Unlock()

	return encoding, nil
}

func (cache *Cache) put(url string, encoding interface{}) {
	strEncoding := fmt.Sprintf("%v", encoding)
	nowTime := time.Now().UnixNano()
	encodingItem := &EncodingItem{strEncoding, nowTime}
	urlItem := &URLItem{url, nowTime}

	cache.Mutex.Lock()
	_, ok := cache.URLToEncodingMap[url]
	if !ok {
		cache.URLToEncodingMap[url] = encodingItem
		cache.EncodingToURLMap[strEncoding] = urlItem
	} else {
		cache.URLToEncodingMap[url].createdAt = nowTime
		cache.EncodingToURLMap[strEncoding].createdAt = nowTime
	}
	cache.Mutex.Unlock()
}

func (cache *Cache) GetEncoding(url string) (string, error) {
	cache.Mutex.Lock()
	item := cache.URLToEncodingMap[url]
	cache.Mutex.Unlock()

	if item != nil {
		return item.encoding, nil
	}

	hash, e := cache.Collection.GetHash(url)
	if e != nil {
		return cache.putWithDBThroughput(url)
	}
	cache.put(url, hash)

	return fmt.Sprintf("%v", hash), nil
}

func (cache *Cache) GetURL(encoding string) (string, error) {
	cache.Mutex.Lock()
	item := cache.EncodingToURLMap[encoding]
	cache.Mutex.Unlock()

	if item != nil {
		return item.url, nil
	}

	url, e := cache.Collection.GetURL(encoding)
	if e != nil {
		return "", e
	}

	if url != "" {
		cache.put(url, encoding)
		return fmt.Sprintf("%v", url), nil
	}

	return "", nil
}
