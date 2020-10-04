package urlshortener

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_encodeAndDecode(t *testing.T) {
	require := require.New(t)
	url := "http://google.com"
	enc := Encode(url)
	decodedURL, err := Decode(enc)

	if err != nil {
		require.Failf("Failed decoding %s ", enc)
	}

	require.Equal(url, decodedURL)
}

func Test_url(t *testing.T) {
	url := "www.google.com"
	enc := Encode(url)
	fmt.Println(enc)
}
