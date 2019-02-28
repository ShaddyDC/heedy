/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package datastream

import (
	"dbsetup/dbutil"
	"os"
	"testing"

	"config"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	log "github.com/Sirupsen/logrus"
)

var (
	sdb *SqlStore
	ds  *DataStream
	mc  *MockCache
	err error
)

//Creates a mocked out cache interface
type MockCache struct {
	mock.Mock
}

func (m *MockCache) StreamLength(deviceID int64, streamID int64, substream string) (int64, error) {
	args := m.Called(deviceID, streamID, substream)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockCache) StreamSize(deviceID int64, streamID int64, substream string) (int64, error) {
	args := m.Called(deviceID, streamID, substream)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockCache) DeviceSize(deviceID int64) (int64, error) {
	args := m.Called(deviceID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCache) Insert(deviceID, streamID int64, substream string, dpa DatapointArray, restamp bool, maxDeviceSize, maxStreamSize int64) (int64, error) {
	args := m.Called(deviceID, streamID, substream, dpa, restamp, maxDeviceSize, maxStreamSize)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockCache) DeleteDevice(deviceID int64) error {
	args := m.Called(deviceID)
	return args.Error(0)
}
func (m *MockCache) DeleteStream(deviceID, streamID int64) error {
	args := m.Called(deviceID, streamID)
	return args.Error(0)
}
func (m *MockCache) DeleteSubstream(deviceID, streamID int64, substream string) error {
	args := m.Called(deviceID, streamID, substream)
	return args.Error(0)
}
func (m *MockCache) ReadProcessingQueue() ([]Batch, error) {
	args := m.Called()
	return args.Get(0).([]Batch), args.Error(1)
}
func (m *MockCache) ReadBatches(batchnumber int) ([]Batch, error) {
	args := m.Called(batchnumber)
	return args.Get(0).([]Batch), args.Error(1)
}
func (m *MockCache) ReadRange(deviceID, streamID int64, substream string, i1, i2 int64) (DatapointArray, int64, int64, error) {
	args := m.Called(deviceID, streamID, substream, i1, i2)
	return args.Get(0).(DatapointArray), args.Get(1).(int64), args.Get(2).(int64), args.Error(3)
}
func (m *MockCache) ClearBatches(b []Batch) error {
	args := m.Called(b)
	return args.Error(0)
}
func (m *MockCache) Close() error {
	return nil
}
func (m *MockCache) Clear() error {
	return nil
}

func TestMain(m *testing.M) {
	mc = &MockCache{}
	sqldb, err := dbutil.OpenDatabase(config.TestConfiguration.Sql.Type, config.TestConfiguration.Sql.GetSqlConnectionString())
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	ds, err = OpenDataStream(mc, sqldb, 2)
	if err != nil {
		log.Error(err)
		os.Exit(3)
	}
	ds.Close()

	ds, err = OpenDataStream(mc, sqldb, 2)
	if err != nil {
		log.Error(err)
		os.Exit(4)
	}
	sdb = ds.sqls

	res := m.Run()

	ds.Close()

	// Once the tests with postgres pass, check sqlite
	if res == 0 {
		mc = &MockCache{}
		err = dbutil.ClearDatabase("sqlite3", "test.db")
		if err != nil {
			if err.Error() != "remove test.db: no such file or directory" {
				panic(err.Error())
			}

		}

		err = dbutil.SetupDatabase("sqlite3", "test.db")
		if err != nil {
			panic(err.Error())
		}
		sqldb, err = dbutil.OpenDatabase("sqlite3", "test.db")
		if err != nil {
			panic(err.Error())
		}

		ds, err = OpenDataStream(mc, sqldb, 2)
		if err != nil {
			log.Error(err)
			os.Exit(3)
		}
		ds.Close()

		ds, err = OpenDataStream(mc, sqldb, 2)
		if err != nil {
			log.Error(err)
			os.Exit(4)
		}
		sdb = ds.sqls

		res = m.Run()

		ds.Close()
	}

	os.Exit(res)
}

func TestBasics(t *testing.T) {
	ds.Clear()

	mc.On("DeleteStream", int64(1), int64(2)).Return(nil)
	require.NoError(t, ds.DeleteStream(1, 2))
	mc.AssertExpectations(t)

	mc.On("StreamLength", int64(1), int64(2), "").Return(int64(0), nil)
	i, err := ds.StreamLength(1, 2, "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)
	mc.AssertExpectations(t)

	mc.On("Insert", int64(1), int64(2), "", dpa6, false, int64(0), int64(0)).Return(int64(5), nil)
	_, err = ds.Insert(1, 2, "", dpa6, false, 0, 0)
	require.NoError(t, err)
	mc.AssertExpectations(t)

	mc.On("DeleteSubstream", int64(1), int64(2), "").Return(nil)
	require.NoError(t, ds.DeleteSubstream(1, 2, ""))

	mc.On("DeleteDevice", int64(1)).Return(nil)
	require.NoError(t, ds.DeleteDevice(1))
}
