package rediscache

import (
	"connectordb/streamdb/datastream"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/require"
)

func TestDataStream(t *testing.T) {
	rc.BatchSize = 2
	sqldb, err := sql.Open("postgres", "sslmode=disable dbname=connectordb port=52592")
	require.NoError(t, err)

	ds, err := datastream.OpenDataStream(RedisCache{rc}, sqldb, 2)
	require.NoError(t, err)

	ds.Clear()

	i, err := ds.StreamLength(0, 1, "")
	require.NoError(t, err)
	require.EqualValues(t, 0, i)

	i, err = ds.Insert(0, 1, "", dpa6, false)
	require.NoError(t, err)
	require.EqualValues(t, 5, i)

	writestrings, err := rc.GetList("BATCHLIST")
	require.NoError(t, err)
	require.Equal(t, 2, len(writestrings))

	//The data was inserted - check if we can get a range from redis
	dr, err := ds.IRange(0, 1, "", 0, 0)
	require.NoError(t, err)
	ar, err := dr.NextArray()
	require.NoError(t, err)
	require.Equal(t, dpa6.String(), ar.String())
	ar, err = dr.NextArray()
	require.NoError(t, err)
	require.Nil(t, ar)

	//Next check if we can get a Trange from the data
	dr, err = ds.TRange(0, 1, "", 1.9, 4.0)
	require.NoError(t, err)
	ar, err = dr.NextArray()
	require.NoError(t, err)
	require.Equal(t, dpa6[1:4].String(), ar.String())
	ar, err = dr.NextArray()
	require.NoError(t, err)
	require.Nil(t, ar)

	//Write the chunks of data
	require.NoError(t, ds.WriteChunk())

	writestrings, err = rc.GetList("BATCHLIST")
	require.NoError(t, err)
	require.Equal(t, 0, len(writestrings))

	//Now check if we can get the range from sql then redis
	dr, err = ds.IRange(0, 1, "", 0, 0)
	require.NoError(t, err)
	ar, err = dr.NextArray()
	require.NoError(t, err)
	require.Equal(t, dpa6[:2].String(), ar.String())
	ar, err = dr.NextArray()
	require.NoError(t, err)
	require.Equal(t, dpa6[2:4].String(), ar.String())
	ar, err = dr.NextArray()
	require.NoError(t, err)
	require.Equal(t, dpa6[4:].String(), ar.String())

	ar, err = dr.NextArray()
	require.NoError(t, err)
	require.Nil(t, ar)

	sqldb.Close()
	rc.BatchSize = 250
}
