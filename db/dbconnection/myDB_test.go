package dbconnection

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDBConnectionCRUD(t *testing.T) {
	collection, mdb := NewTestDBCollectionConnection(t)
	//linter: errcheck
	defer mdb.Terminate(nil)

	url := "aaa"
	hash := "333"

	insert, e := collection.Insert(url, hash)
	require := require.New(t)
	require.NoError(e)
	require.NotNil(insert)
	require.NotZero(insert)

	h, e := collection.GetHash(url)
	require.NoError(e)
	require.Equal(hash, h)

	s, e := collection.GetURL(hash)
	require.NoError(e)
	require.Equal(url, s)

	ok, e := collection.Delete(url)
	require.NoError(e)
	require.True(ok)

	ok, e = collection.Delete(url)
	require.NoError(e)
	require.False(ok)

	t.Logf("Terminating the mongo db test container")
}
