//Dtypes is a package that handles data types for timebatchdb. It allows saving and converting between different data
//types transparently to timebatchdb's byte arrays
package dtypes

import (
	"database/sql"
	"errors"
	"log"
	"streamdb/timebatchdb"
)

var (
	ERROR_KEYNOTFOUND = errors.New("Key not found in datapoint")

//ERROR_UNKNOWNDTYPE = errors.New("Unrecognized data type")
)

//A simple wrapper for DataRange which returns marshalled data
type TypedRange struct {
	dr    timebatchdb.DataRange
	dtype DataType
}

func (tr TypedRange) Close() {
	tr.dr.Close()
}
func (tr TypedRange) Next() TypedDatapoint {
	d, err := tr.dr.Next()
	if err != nil {
		log.Printf("Error in TypedRange: %s", err)
	}
	if d == nil {
		return nil
	}
	dp := tr.dtype.New()
	err = dp.Load(*d)
	if err != nil {
		return nil
	}
	return dp
}

//Simple type wrapper for timebatchdb's Database
type TypedDatabase struct {
	db *timebatchdb.Database
}

func (d *TypedDatabase) Close() {
	d.db.Close()
}

//Returns the DataRange associated with the given time range
func (d *TypedDatabase) GetTimeRange(key string, dtype string, starttime int64, endtime int64) TypedRange {
	t, ok := GetType(dtype)
	if !ok {
		log.Printf("TypedDatabase.Get: Unrecognized type '%s'\n", dtype)
		return TypedRange{timebatchdb.EmptyRange{}, NilType{}}
	}
	log.Printf("Requesting timestamp data (%v, %v] from '%v'", starttime, endtime, key)
	val, err := d.db.GetTimeRange(key, starttime, endtime)
	if err != nil {
		log.Printf("Error getting by timestamps: %s", err)
	}
	return TypedRange{val, t}
}

//Returns the DataRange associated with the given index range
func (d *TypedDatabase) GetIndexRange(key string, dtype string, startindex uint64, endindex uint64) TypedRange {
	t, ok := GetType(dtype)
	if !ok {
		log.Printf("TypedDatabase.Get: Unrecognized type '%s'\n", dtype)
		return TypedRange{timebatchdb.EmptyRange{}, NilType{}}
	}
	log.Printf("Requesting data (%v, %v] from '%v'", startindex, endindex, key)
	val, err := d.db.GetIndexRange(key, startindex, endindex)
	if err != nil {
		log.Printf("Error getting index: %s", err)
	}
	return TypedRange{val, t}
}

//Inserts the given data into the DataStore, and uses the given routing address for data
func (d *TypedDatabase) Insert(datapoint TypedDatapoint) error {
	s := datapoint.Key()
	if s == "" {
		return ERROR_KEYNOTFOUND
	}
	return d.InsertKey(s, datapoint)
}
func (d *TypedDatabase) InsertKey(key string, datapoint TypedDatapoint) error {
	log.Printf("Inserting: '%s'\n", key)
	timestamp, err := datapoint.Timestamp()
	data := datapoint.Data()
	if err != nil {
		return err
	}
	dpa := timebatchdb.CreateDatapointArray([]int64{timestamp}, [][]byte{data})
	return d.db.Insert(key, dpa)
}

//Opens the DataStore.
func Open(sdb *sql.DB, sqlstring string, redisurl string, batchsize int, err error) (*TypedDatabase, error) {
	var td TypedDatabase
	err = td.InitTypedDB(sdb, sqlstring, redisurl, batchsize, err)

	if err != nil {
		return nil, err
	}

	return &td, nil
	/**
	  TODO removeme when all tests check out.

	  ds,err := timebatchdb.Open(msgurl,mongourl,mongoname)
	  if err!=nil {
	      return nil,err
	  }
	  return &TypedDatabase{ds},nil
	  **/
}

// Initializes a Typed Database that already exists.
func (d *TypedDatabase) InitTypedDB(sdb *sql.DB, sqlstring string, redisurl string, batchsize int, err error) error {
	ds, err := timebatchdb.Open(sdb, sqlstring, redisurl, batchsize, err)
	if err != nil {
		return err
	}
	d.db = ds
	return nil
}

func (d *TypedDatabase) WriteDatabase() (err error) {
	return d.db.WriteDatabase()
}
