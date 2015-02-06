package timebatchdb

import (
    "bytes"
    "container/list"
    "compress/gzip"
    )

type DatapointArray struct {
    Datapoints []Datapoint          //The array of datapoints
    array []byte                    //A (possibly nil) single byte array which holds all of the datapoints
    iloc int                        //Allows to make this a DataRange object
}

//The number of datapoints contained in the page
func (d *DatapointArray) Len() int {
    return len(d.Datapoints)
}

//Total size of the page in bytes (uncompressed)
func (d *DatapointArray) Size() int {
    if (d.array!=nil) {
        return len(d.array)
    }
    size:=0
    for i:=0;i<len(d.Datapoints);i++ {
        size+= d.Datapoints[i].Len()
    }
    return size
}

//Dummy function - allows datapointArray to conform to the DataRange interface
func (d *DatapointArray) Init() {

}
//Dummy function - allows datapointArray to conform to the DataRange interface.
//This doesn't actually do anything, and you don't need to call this
func (d *DatapointArray) Close() {

}

//Allows to use DatapointArray as a DataRange - starts from the first datapoint, and
//successively returns datapoint ptrs until there are none left, at which point it returns nil.
//It is an iterator
func (d *DatapointArray) Next() *Datapoint {
    if d.iloc >= d.Len() {
        return nil
    }
    dp := &d.Datapoints[d.iloc]
    d.iloc++
    return dp
}

//Resets the iterator back to 0
func (d *DatapointArray) Reset() {
    d.iloc = 0
}

//Returns the timestamps associated with the index range
func (d *DatapointArray) TimestampIRange(start int, end int) (timestamps []uint64) {
    if (end > d.Len()) {
        end = d.Len()
    }
    if (start > end) {
        return nil
    }
    timestamps = make([]uint64,end-start)
    for i:=start;i<end;i++ {
        timestamps[i] = d.Datapoints[i].Timestamp()
    }
    return timestamps
}

//Returns the array of timestamps
func (d *DatapointArray) Timestamps() (timestamps []uint64) {
    return d.TimestampIRange(0,d.Len())
}

//Returns the data associated with the index range
func (d *DatapointArray) DataIRange(start int, end int) (data [][]byte) {
    if (end > d.Len()) {
        end = d.Len()
    }
    if (start > end) {
        return nil
    }
    data = make([][]byte,end-start)
    for i:=start;i<end;i++ {
        data[i] = d.Datapoints[i].Data()
    }
    return data
}

//Returns the array of data
func (d *DatapointArray) Data() (data [][]byte) {
    return d.DataIRange(0,d.Len())
}

//Returns the array of timestamps and the array of associated data
func (d *DatapointArray) Get() (timestamps []uint64, data [][]byte) {
    return d.Timestamps(),d.Data()
}

//Find the index of the first datapoint in the array which has a timestamp strictly greater
//than the given reference timestamp.
//If no datapoints fit this, returns -1
//(ie, no datapoint in array has a timestamp greater than the given time)
func (d *DatapointArray) FindTimeIndex(timestamp uint64) int {
    //BUG(daniel): This code makes no guarantees about nanosecond-level precision.
    if (d.Len()==0) {
        return -1
    }

    leftbound := 0
    leftts := d.Datapoints[0].Timestamp()

    //If the timestamp is earlier than the earliest datapoint
    if (leftts > timestamp) {
        return 0
    }

    rightbound := d.Len()-1                        //Len is guaranteed > 0
    rightts := d.Datapoints[rightbound].Timestamp()


    if (rightts <= timestamp) {
        return -1
    }

    //We do this shit logn style
    for (rightbound - leftbound > 1) {
        midpoint := (leftbound + rightbound)/2
        ts := d.Datapoints[midpoint].Timestamp()
        if (ts <= timestamp) {
            leftbound = midpoint
            leftts = ts
        } else {
            rightbound = midpoint
            rightts = ts
        }
    }
    return rightbound
}

//Returns a DatapointArray which has the given starting bound (like DatapointTRange, but without upperbound)
func (d *DatapointArray) TStart(timestamp uint64) *DatapointArray {
    i := d.FindTimeIndex(timestamp)
    if i==-1 {
        return nil
    }
    return  NewDatapointArray(d.Datapoints[i:])
}

//Returns the DatapointArray of datapoints which fit within the time range:
//  (timestamp1,timestamp2]
func (d *DatapointArray) DatapointTRange(timestamp1 uint64, timestamp2 uint64) *DatapointArray {
    i1 := d.FindTimeIndex(timestamp1)
    if i1==-1 {
        return nil
    }
    i2 := d.FindTimeIndex(timestamp2)
    if i2==-1 {
        //The endrange is out of bounds - read until the end of the array
        return  NewDatapointArray(d.Datapoints[i1:])
    }
    return  NewDatapointArray(d.Datapoints[i1:i2])
}

func (d *DatapointArray) DataTRange(timestamp1 uint64, timestamp2 uint64) [][]byte {
    return d.DatapointTRange(timestamp1,timestamp2).Data()
}
func (d *DatapointArray) TimestampTRange(timestamp1 uint64, timestamp2 uint64) []uint64 {
    return d.DatapointTRange(timestamp1,timestamp2).Timestamps()
}
//Returns the array of timestamps and data which fit in the given time range:
//  (timestamp1,timestamp2]
func (d *DatapointArray) GetTRange(timestamp1 uint64, timestamp2 uint64) ([]uint64,[][]byte) {
    return d.DatapointTRange(timestamp1,timestamp2).Get()
}


//Returns the byte array associated with the entire page of datapoints (uncompressed)
func (d *DatapointArray) Bytes() []byte {
    if (d.array!=nil) {
        return d.array
    }
    //The array does not exist. We therefore create it
    arr := make([]byte,d.Size())
    n := 0
    for i:=0;i<len(d.Datapoints);i++ {
        num_written := copy(arr[n:],d.Datapoints[i].Bytes())
        //In the interest of saving memory, have the datapoints refer to slices of the newly created
        //byte array, rather than having multiple copies of the same data
        d.Datapoints[i] = Datapoint{arr[n:n+num_written]}
        n+= num_written
    }
    d.array = arr
    return arr
}

//Returns the gzipped bytes of the entire page of datapoints
func (d *DatapointArray) CompressedBytes() []byte {
    var b bytes.Buffer
    w := gzip.NewWriter(&b)
    w.Write(d.Bytes())
    w.Close()
    return b.Bytes()
}


//Creates a DatapointArray given an actual datapoint array
func NewDatapointArray(d []Datapoint) *DatapointArray {
    return &DatapointArray{d,nil,0}
}

//Creates DatapointArray from the raw stuff
func CreateDatapointArray(timestamps []uint64, data [][]byte) *DatapointArray {
    arr := make([]Datapoint,len(timestamps))

    for i:=0;i<len(arr);i++ {
        arr[i] = NewDatapoint(timestamps[i],data[i])
    }
    return NewDatapointArray(arr)
}

//Creates a datapoint array from its associated bytes. Note that the Datapoint array assumes
//that the bytes are correctly sized, unlike the Datapoint and KeyedDatapoint functions.
func DatapointArrayFromBytes(data []byte) *DatapointArray {
    if (len(data) ==0) {
        return nil
    }
    n := 0
    l := list.New()
    for n < len(data) {
        dp,num := DatapointFromBytes(data[n:])
        l.PushBack(dp)
        n += int(num)
    }

    dp := DatapointArrayFromList(l)
    dp.array = data[:n]    //We can set the array too

    return dp
}

//Given a linked list, create the DatapointArray
func DatapointArrayFromList(l *list.List) *DatapointArray {
    //Now create the points array
    points := make([]Datapoint,l.Len())
    elem:=l.Front()
    points[0] = elem.Value.(Datapoint)
    for j:=1;elem.Next()!=nil;j++ {
        elem = elem.Next()
        points[j] = elem.Value.(Datapoint)
    }

    return &DatapointArray{points,nil,0}
}

//Given the correctly sized byte array for the compressed representation of a DatapointArray,
//decompress it.
func DatapointArrayFromCompressedBytes(cdata []byte) *DatapointArray {
    r, _ := gzip.NewReader(bytes.NewBuffer(cdata))

    l := list.New()

    d,err := ReadDatapoint(r)
    for err==nil {
        l.PushBack(d)
        d,err = ReadDatapoint(r)
    }
    r.Close()

    return DatapointArrayFromList(l)
}


//Given a DataRange, creates a DatapointArray based upon it. Closes the DataRAnge when done
func DatapointArrayFromDataRange(dr DataRange) *DatapointArray {
    l := list.New()

    d := dr.Next()
    for d!=nil {
        l.PushBack(*d)
        d = dr.Next()
    }
    dr.Close()

    return DatapointArrayFromList(l)
}
