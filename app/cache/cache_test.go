package cache

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go-tiny-url-toy-system/db/dbconnection"
)

func TestCache_CreateRead(t *testing.T) {
	require := require.New(t)
	url := "aaa"

	collection, mdb := dbconnection.NewTestDBCollectionConnection(t)
	defer mdb.Terminate(context.Background())

	myCache := Cache{}
	if err4 := myCache.Init(collection); err4 != nil {
		log.Fatalf("Failed to init cache %+v", err4)
	}

	encoding, err := myCache.GetEncoding(url)
	require.NoError(err)
	require.NotNil(encoding)
	require.NotEmpty(encoding)

	encoding1, err1 := collection.GetHash(url)
	require.NoError(err1)
	require.Equal(encoding, encoding1)

	mappedURL, err2 := collection.GetURL(encoding)
	require.NoError(err2)
	require.Equal(url, mappedURL)

	log.Println("Checking cache has been cleared")
	time.Sleep(time.Duration(ImaxTTL))
	log.Println("Woke up")
	encoding3, ok := myCache.URLToEncodingMap[url]
	require.False(ok)
	require.Nil(encoding3)
}
