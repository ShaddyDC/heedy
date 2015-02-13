package datastore

import (
    "testing"
    "time"
    )

func TestDataStore(t *testing.T) {
    //Turn on the DataStore writer
    go DataStoreWriter("localhost:4222","localhost","testdb", "testing/>")

    m,err := OpenMongoStore("localhost","testdb")
    if (err!=nil) {
       t.Errorf("Couldn't open MongoStore")
       return
    }
    defer m.Close()

    //First drop the collection - so that tests are fresh
    m.DropCollection("0")

    db,err := Open("localhost:4222","localhost","testdb")
    if err!=nil {
        t.Errorf("Couldn't connect: %s",err)
        return
    }
    defer db.Close()

    //Wait one second for the DataStoreWriter to initialize
    time.Sleep(500 * time.Millisecond)

    timestamps := []int64{1,2,3,4,5,6,3000,3100,3200}
    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8")}

    for i:=0;i<len(timestamps);i++ {
        err = db.Insert("user1/device1/stream1",timestamps[i],data[i],"testing/test")
        if err!=nil {
            t.Errorf("Insert Failed: %s",err)
        }
    }

    //Wait one second for the datapoints to be committed to the DataStore
    time.Sleep(500 * time.Millisecond)

    //Now check a data range by index, and then by timestamp
    r := db.GetTimeRange("user1/device1/stream2",0,1000)
    defer r.Close()
    dp:= r.Next()
    if (dp!=nil) {
        t.Errorf("Insert wrong key")
        return
    }

    //Now check a data range by index, and then by timestamp
    r = db.GetIndexRange("user1/device1/stream2",0,1000)
    defer r.Close()
    dp = r.Next()
    if (dp!=nil) {
        t.Errorf("Insert wrong key")
        return
    }

    r = db.GetTimeRange("user1/device1/stream1",2,5)
    defer r.Close()
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=3) {
        t.Errorf("Insert wrong key")
        return
    }
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=4) {
        t.Errorf("Insert wrong key")
        return
    }
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=5) {
        t.Errorf("Insert wrong key")
        return
    }
    dp= r.Next()
    if (dp!=nil) {
        t.Errorf("Insert wrong key")
        return
    }

    r = db.GetIndexRange("user1/device1/stream1",2,5)
    defer r.Close()
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=3) {
        t.Errorf("Insert wrong key")
        return
    }
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=4) {
        t.Errorf("Insert wrong key")
        return
    }
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=5) {
        t.Errorf("Insert wrong key")
        return
    }
    dp= r.Next()
    if (dp!=nil) {
        t.Errorf("Insert wrong key")
        return
    }

}
