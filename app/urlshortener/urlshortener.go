package urlshortener

import (
	b64 "encoding/base64"
)

func Encode(url string) string {
	return b64.URLEncoding.EncodeToString([]byte(url))
}

func Decode(enc string) (string, error) {
	bytes, e := b64.URLEncoding.DecodeString(enc)
	return string(bytes), e
}
