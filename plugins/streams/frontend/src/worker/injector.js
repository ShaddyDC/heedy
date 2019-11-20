import DataManager from "./datamanager.js";

class StreamInjector {
    constructor(wkr) {
        this.worker = wkr;
        this.processors = {};

        this.streams = {};

        wkr.addHandler("stream_query", (ctx, msg) => this._query(ctx, msg));
        wkr.addHandler("stream_subscribe_query", (ctx, msg) => this._subscribeQuery(ctx, msg));
        wkr.addHandler("stream_unsubscribe_query", (ctx, msg) => this._unsubscribeQuery(ctx, msg));

        // TODO: In the future, make sure to subscribe to streams from other users
        // that might be queried
        if (wkr.info.user != null) {
            wkr.websocket.subscribe("stream_data_write", {
                event: "stream_data_write",
                user: wkr.info.user.username
            }, (e) => this._dataEvent(e));
            wkr.websocket.subscribe("stream_actions_write", {
                event: "stream_actions_write",
                user: wkr.info.user.username
            }, (e) => this._dataEvent(e));
            /* object updates happen through re-subscribing
            wkr.websocket.subscribe("object_update_streamdata", {
                event: "object_update",
                user: wkr.info.user.username
            }, (e) => this._objectEvent(e));
            */
            wkr.websocket.subscribe("object_delete_streamdata", {
                event: "object_delete",
                user: wkr.info.user.username
            }, (e) => this._objectEvent(e));

            // In a perfect world, we would also subscribe to object_update. 
            // However, having the streams come up from the frontend instead allows
            // us to avoid an API query - otherwise each time the object is updated,
            // there would be 2 queries, one from the frontend, and one from the worker.
            // This way, the frontend queries, and the worker gets the results of that query.
            wkr.addHandler("stream_update", (ctx, msg) => this._streamUpdate(msg));

            wkr.websocket.subscribe_status((s) => this._ws_status(s));
        }

    }

    addProcessor(key, f) {
        this.processors[key] = f;
    }

    _ws_status(s) {
        Object.values(this.streams).forEach(sv => sv.onWebsocket(s));
    }

    async _dataEvent(event) {
        console.log("stream_worker: DATA EVENT", event);
        if (this.streams[event.object] !== undefined) {
            this.streams[event.object].onEvent(event);
        }
    }
    async _objectEvent(event) {
        console.log("stream_worker: object event", event);
        if (this.streams[event.object] !== undefined) {
            if (event.event == "object_delete") {
                this.streams[event.object].clear();
                delete this.streams[event.object];
            }
        }

    }
    async _streamUpdate(stream) {
        if (this.streams[stream.id] !== undefined) {
            this.streams[stream.id].updateStream(stream);
        }
    }
    async _subscribeQuery(ctx, msg) {
        if (this.streams[msg.stream.id] === undefined) {
            this.streams[msg.stream.id] = new DataManager(this, msg.stream);
        }
        this.streams[msg.stream.id].subscribe(msg.stream, msg.key, msg.query);
    }
    async _unsubscribeQuery(ctx, msg) {
        if (this.streams[msg.id] !== undefined) {
            this.streams[msg.id].unsubscribe(msg.key);
        }

    }
    async _query(ctx, msg) {
        if (this.streams[msg.stream.id] === undefined) {
            this.streams[msg.stream.id] = new DataManager(this, msg.stream);
        }
        this.streams[msg.stream.id].query(msg.stream, msg.key, msg.query);
    }


}

export default StreamInjector;