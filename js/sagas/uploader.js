import { delay } from 'redux-saga'
import { put, select, takeLatest } from 'redux-saga/effects'
import moment from 'moment';

import storage from '../storage';
import { ConnectorDB } from 'connectordb';
import { cdbPromise } from '../util';
import { go } from '../actions';

function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

function* process(action) {
    let uploader = yield select((state) => state.pages.uploader);
    let txt = uploader.part1.rawdata;
    let transform = uploader.part2.transform.trim();
    // We have the text of the data. Let's try to process it
    let data = [];

    // Clear the error
    yield put({ type: "UPLOADER_PART2", value: { error: "" } });
    try {
        data = JSON.parse(txt);
        // Holy crap, the data was JSON...
    } catch (e) {
        // The data is NOT JSON. We assume it is CSV, and let PapaParse do its magic :)
        let result = Papa.parse(txt, { skipEmptyLines: true, dynamicTyping: true, header: true });
        if (result.errors.length > 0) {
            console.log("Parsing Error:", result);
            yield put({ type: "UPLOADER_PART2", value: { error: result.errors[0].message } });
            return;
        }
        data = result.data;
    }
    console.log("Data before processing", data);

    if (data.length == 0) {
        yield put({ type: "UPLOADER_PART2", value: { error: "No data found" } });
    }

    let d = data[0];
    let keys = Object.keys(d);
    if (keys.length <= 1) {
        yield put({ type: "UPLOADER_PART2", value: { error: "Datapoints need to have timestamp and data fields" } });
        return;
    }

    // First, let's process the field names, so they are cleaned up. We remove capitalized fields,
    // since by default, fields are lowercase in ConnectorDB
    // We also start looking for timestamps in order:
    // If 'timestamp' is a field, and can be parsed as a timestamp, it is utomatically chosen as 'the' timestamp field
    // if 't' is a field, and can be parsed as either a timestamp or a unix time, it is chosen

    let fieldMap = {};
    let hadTimestamp = "";
    let hadT = "";
    for (let i = 0; i < keys.length; i++) {
        let curkey = keys[i];
        fieldMap[curkey] = curkey.trim().replace(/\s+/g, '_').toLowerCase();
        if (fieldMap[curkey] === "timestamp" && moment(d[curkey]).isValid()) {
            hadTimestamp = curkey;
        }
        if (fieldMap[curkey] === "t" && moment(d[curkey]).isValid()) {
            hadT = curkey;
        }
    }
    let ts = "";
    if (hadTimestamp !== "") {
        ts = hadTimestamp;
    } else if (hadT !== "") {
        ts = hadT;
    } else {
        // So there was no obvious timestamp field. Shame. We now look for a field that 
        // can be parsed as a timestamp. We first check for a field that can be parsed as a STRING,
        // since that means it is super likely to be a timestamp.
        let tsfields = {};
        for (let i = 0; i < keys.length; i++) {
            if (typeof d[keys[i]] === 'string') {
                let v = moment(d[keys[i]]);
                if (v.isValid()) {
                    tsfields[keys[i]] = v;
                }
            }
        }

        if (Object.keys(tsfields).length != 0) {
            // We found at least one timestamp field. We just return the first one.
            // Because if the user has multiple, they can just set the field name to "timestamp"
            // to force it to be chosen
            ts = Object.keys(tsfields)[0];
        } else {
            // No string timestamp fields. We now look for unix timestamps. We choose the best
            // one by its proximity to 2018 - data that isn't timestamps usually won't
            // be the closest number
            let best = 0;
            let bestKey = "";
            for (let i = 0; i < keys.length; i++) {
                if (typeof d[keys[i]] === 'number') {
                    if (Math.abs(d[keys[i]] - 1514786400) < Math.abs(best - 1514786400)) {
                        best = d[keys[i]];
                        bestKey = keys[i];
                    }
                }
            }

            if (bestKey === "") {
                yield put({ type: "UPLOADER_PART2", value: { error: "Could not find timestamps" } });
                return;
            }

            ts = bestKey;
        }
    }

    console.log("Using field " + ts + " as timestamp for processing data");

    // OK, at this point, ts is the key of the timestamp. We create a function that will automatically
    // convert the data into a unix timestamp given a datapoint 
    let getT = (dp) => dp[ts];
    if (typeof d[ts] === 'string') {
        getT = (dp) => moment(dp[ts]).unix();
    }

    // Now we create a generator for the data portion of the datapoint. We remove the timestamp's key from
    // the dataset, so it is not included in the resulting dataset.
    keys.splice(keys.indexOf(ts), 1);

    let getD = (dp) => dp[keys[0]];

    if (keys.length > 1) {
        getD = (dp) => {
            let result = {};
            for (let i = 0; i < keys.length; i++) {
                result[fieldMap[keys[i]]] = dp[keys[i]];
            }
            return result;
        };
    }

    // ... and we finally generate the Datapoints
    let resultingData = new Array(data.length);
    let j = 0;
    for (let i = 0; i < data.length; i++) {
        try {
            resultingData[j] = {
                t: getT(data[i]),
                d: getD(data[i])
            };
            j++;
        } catch (e) {
            console.log("Got error, skipping datapoint", e);
        }
    }

    // We name the result data.
    data = resultingData.slice(0, j);

    // Finally, let's make sure it is sorted by timestamp
    data.sort((a, b) => (a.t > b.t ? 1 : (a.t < b.t ? -1 : 0)));

    console.log("Finished processing data", data);

    // Finally, let's run the given transform on the data
    if (transform !== "") {
        try {
            data = pipescript.Script(transform).Transform(data);
        } catch (e) {
            yield put({ type: "UPLOADER_PART2", value: { error: e.toString() } });
            return;
        }
        console.log("After transform", transform, data);
    }


    yield put({ type: "UPLOADER_SET", value: { data: data } });

}

function* upload(action) {
    let uploader = yield select((state) => state.pages.uploader);
    let username = yield select((state) => state.site.thisUser.name);

    // Clear the error
    yield put({ type: "UPLOADER_PART3", value: { error: "" } });

    if (uploader.data.length === 0) {
        yield put({ type: "UPLOADER_PART3", value: { error: "Process the data first." } });
        return;
    }
    let s = uploader.part3.stream.split("/");
    let data = uploader.data;
    if (uploader.part3.stream.length === 0 || s.length != 3 || s[0].length == 0 || s[1].length == 0 || s[2].length == 0) {
        yield put({ type: "UPLOADER_PART3", value: { error: "Invalid stream name" } });
        return;
    }

    if (s[0] !== username) {
        yield put({ type: "UPLOADER_PART3", value: { error: "Can't upload to other users" } });
        return;
    }

    // Show loading bar
    yield put({ type: "UPLOADER_PART3", value: { loading: true, percentdone: 0 } });

    // Now let's try to read the device that will own the stream
    let device = null;
    try {
        device = (yield cdbPromise(storage.cdb.readDevice(username, s[1]), 5 * 1000));
    } catch (e) {
        if (e.toString() != "Error: Can't access this resource.") {
            console.log(e.toString());
            yield put({ type: "UPLOADER_PART3", value: { error: e.toString(), loading: false } });
            return;
        }
        if (s[1] !== "uploads") {
            console.log("Looks like the device doesn't exist.");
            yield put({ type: "UPLOADER_PART3", value: { error: "The chosen device does not exist", loading: false } });
            return;
        }
        console.log("The 'uploads' device doesn't exist. Creating it...");
        try {
            device = (yield cdbPromise(storage.cdb.createDevice(username, {
                name: "uploads",
                icon: "material:file_upload",
                description: "Device for holding manually uploaded data"
            }), 5 * 1000));
        } catch (e) {
            console.log(e.toString());
            yield put({ type: "UPLOADER_PART3", value: { error: e.toString(), loading: false } });
            return;
        }
    }

    console.log("Have device", device);

    // We now log in as the device, so that we can upload data to its streams
    let cdb = new ConnectorDB(device.apikey, undefined, "//" + window.location.host, false);

    // Check to see if the login was successful
    try {
        let ping = yield cdbPromise(cdb._doRequest("?q=this", "GET"), 5 * 1000);
        if (ping !== s[0] + "/" + s[1]) {
            console.log("Ping result:", ping);
            yield put({ type: "UPLOADER_PART3", value: { error: "Unable to log in as " + s[0] + "/" + s[1], loading: false } });
            return;
        }
    } catch (e) {
        console.log(e.toString());
        yield put({ type: "UPLOADER_PART3", value: { error: e.toString(), loading: false } });
        return;
    }

    // We are logged in! Check if the stream exists
    try {
        let stream = (yield cdbPromise(cdb.readStream(s[0], s[1], s[2]), 5 * 1000));

        // The stream exists!
        if (!uploader.part3.overwrite) {
            yield put({ type: "UPLOADER_PART3", value: { error: "The stream already exists", loading: false } });
            return;
        }

        // Since the stream exists, we need to make sure that the data timestamps don't interfere.
        // Check the last datapoint
        try {
            let existingdata = (yield cdbPromise(cdb.indexStream(s[0], s[1], s[2], -1, 0), 5 * 1000));
            if (existingdata.length > 0) {
                // We have a datapoint. We check if the timestamp is newer than a point in our current dataset
                if (data[0].t < existingdata[0].t) {
                    console.log("Newer data already exists in stream " + data[0].t + " < " + existingdata[0].t);
                    if (!uploader.part3.removeolder) {
                        yield put({ type: "UPLOADER_PART3", value: { error: "Newer data already exists in stream", loading: false } });
                        return;
                    }

                    // We now clear the older datapoints from our to-upload Array
                    let i = 0;
                    for (i = 0; i < data.length; i++) {
                        if (data[i].t > existingdata[0].t) {
                            break;
                        }
                    }
                    if (data.length == i) {
                        yield put({ type: "UPLOADER_PART3", value: { error: "Data existing in stream is newer than all datapoints.", loading: false } });
                        return;
                    }

                    // Now we slice the data array not to include any old datapoints
                    data = data.slice(i);
                    console.log("After clearing " + i + " datapoints, left with", data);
                }
            }
        } catch (e) {
            console.log(e.toString());
            yield put({ type: "UPLOADER_PART3", value: { error: e.toString(), loading: false } });
            return;
        }


    } catch (e) {
        if (e.toString() != "Error: Can't access this resource.") {
            console.log(e.toString());
            yield put({ type: "UPLOADER_PART3", value: { error: e.toString(), loading: false } });
            return;
        }
        if (!uploader.part3.create) {
            yield put({ type: "UPLOADER_PART3", value: { error: "The stream doesn't exist", loading: false } });
            return;
        }

        // Create the stream
        try {
            yield cdbPromise(cdb.createStream(username, s[1], {
                name: s[2],
                icon: "material:file_upload"
            }), 5 * 1000);
        } catch (e) {
            console.log(e.toString());
            yield put({ type: "UPLOADER_PART3", value: { error: e.toString(), loading: false } });
            return;
        }
    }

    // We're ready to start inserting the data! Whew!
    // We insert it in batches of 5000, because we don't want to hit upon the insert size limit of ConnectorDB
    let total = data.length;
    let path = uploader.part3.stream;
    let batch = 1000;
    try {
        while (data.length > 0) {
            if (data.length <= batch) {
                yield cdbPromise(cdb._doRequest("crud/" + path + "/data", "POST", data));
                data = [];
            } else {
                yield cdbPromise(cdb._doRequest("crud/" + path + "/data", "POST", data.slice(0, batch)));
                data = data.slice(batch);
            }
            yield put({ type: "UPLOADER_PART3", value: { percentdone: (total - data.length) * 100 / total } });
        }
    } catch (e) {
        console.log(e.toString());
        yield put({ type: "UPLOADER_PART3", value: { error: e.toString(), loading: false } });
        return;
    }

    // Go to the stream
    yield put(go(path));

    // Clear the loading bar
    yield put({ type: "UPLOADER_PART3", value: { loading: false, percentdone: 0 } });




}

// We automatically preset the stream name to use the uploads device
function* navigate(action) {
    if (action.payload.hash !== "#upload" || action.payload.pathname !== "/") {
        return;
    }
    let uploader = yield select((state) => state.pages.uploader);
    if (uploader.part3.stream === "") {
        let username = yield select((state) => state.site.thisUser.name);
        yield put({ type: "UPLOADER_PART3", value: { stream: username + "/uploads/" } });
    }

}

// Our watcher Saga: spawn a new incrementAsync task on each INCREMENT_ASYNC
export default function* uploaderSaga() {
    yield takeLatest('UPLOADER_PROCESS', process);
    yield takeLatest('UPLOADER_UPLOAD', upload);
    yield takeLatest('@@router/LOCATION_CHANGE', navigate);
}