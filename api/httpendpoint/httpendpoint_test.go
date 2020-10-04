package httpendpoint

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go-tiny-url-toy-system/db/dbconnection"
)

const url = "www.google.com"
const expectedEncoding = "d3d3Lmdvb2dsZS5jb20="
const expectedBody = "<a href=\"/www.google.com\">See Other</a>.\n\n"

func TestGetShortenAndRedirect(t *testing.T) {
	collection, mdb := dbconnection.NewTestDBCollectionConnection(t)
	//nolint:errcheck
	defer mdb.Terminate(nil)
	s := NewServer()

	s.Start(collection)
	defer s.Shutdown()

	require := require.New(t)

	req, err := http.NewRequest("GET", fmt.Sprintf("/shorten?url=%s", url), nil)
	require.NoError(err)
	rr := httptest.NewRecorder()
	s.HandleShorten().ServeHTTP(rr, req)
	status := rr.Code
	require.Equal(http.StatusCreated, status)
	require.Equal(expectedEncoding, rr.Body.String())

	req, err = http.NewRequest("GET", fmt.Sprintf("/redirect?hash=%s", expectedEncoding), nil)
	require.NoError(err)
	rr = httptest.NewRecorder()
	s.HandleRedirect().ServeHTTP(rr, req)
	status = rr.Code
	require.Equal(http.StatusSeeOther, status)
	require.Equal(expectedBody, rr.Body.String())
}
