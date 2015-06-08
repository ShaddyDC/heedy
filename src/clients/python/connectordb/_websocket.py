import websocket
import threading
import logging
import json
import errors
import random

class WebsocketHandler(object):
    #SubscriptionHandler manages the websocket connection and fires off all subscriptions
    def __init__(self,url,basicAuth):
        self.uri = self.getWebsocketURI(url)
        self.headers = self.getAuthHeaders(basicAuth)

        self.subscriptions = {}
        self.subscription_lock = threading.Lock()

        self.isconnected = False
        

        self.ws = None
        self.ws_thread = None
        self.ws_openlock = threading.Lock()    #Allows us to synchronously wait for connection to be ready
        self.ws_sendlock = threading.Lock()

        #If it wants a connection, then if websocket dies, it is reconnected immediately
        self.wantsconnection = False
        self.isretry = False
        self.reconnectbackoff= 1.0


    def getWebsocketURI(self,url):
        #Given a URL to the REST API, get the websocket uri
        ws_url = "wss://" + url[8:]
        if url.startswith("http://"):   #Unsecured websocket is only really there for testing
            ws_url = "ws://"+ url[7:]
        return ws_url

    def getAuthHeaders(self,basicAuth):
        #Use a cheap hack to extract the basic auth header from a requests HTTPBasicAuth object
        class tmpObject():
            def __init__(self):
                self.headers = {}
        tobj = tmpObject()
        basicAuth(tobj)

        headers = []
        for header in tobj.headers:
            headers.append("%s: %s"%(header,tobj.headers[header]))
        return headers

    def unlockopen(self,isconnected=False):
        #Unlocks the open
        self.isconnected = isconnected
        try:
            self.ws_openlock.release()
            return True
        except:
            return False


    def __on_message(self,ws,msg):
        msg= json.loads(msg)
        logging.debug("ConnectorDB: Got message for '%s'",msg["stream"])

        #Alright - now that we have the data, we call callbacks, in order:
        self.subscription_lock.acquire()

        def runfnc(fnc):
            #Run the callbacks, but release the lock, just in case subscribing/unsubscribing happens
            #in the callbacks
            self.subscription_lock.release()
            s = msg["stream"]
            res = fnc(s,msg["data"])
            if res==True:
                #This is a convenience function - if True is returned by the subcription callback,
                #it means that the callback wants the same datapoints (without changed timestamps) to be inserted.
                res = msg["data"]
            if res!=False and res is not None and s.endswith("/downlink") and s.count("/")==3:
                #The downlink was acknowledged - write the datapoints through websocket,
                #so that it is visible that it was processed
                self.insert(msg["stream"][:-9],res)
            self.subscription_lock.acquire()

        if msg["stream"] in self.subscriptions:
            runfnc(self.subscriptions[msg["stream"]])
            

        #Now handle the more general subscriptions to device or user
        pathparts = msg["stream"].split("/")

        if len(pathparts)==3:
            #We don't want to get downlinks or substreams in this
            if pathparts[0] in self.subscriptions:
                runfnc(self.subscriptions[pathparts[0]])
              
            if pathparts[0]+"/"+pathparts[1] in self.subscriptions:
                runfnc(self.subscriptions[pathparts[0]+"/"+pathparts[1]])

        self.subscription_lock.release()

    def __on_open(self,ws):
        logging.debug("ConnectorDB: Websocket opened")
        self.unlockopen(True)
        self.reconnectbackoff=1.0
    def __on_close(self,ws):
        logging.debug("ConnectorDB: Websocket Closed")
        self.unlockopen()
    def __on_error(self,ws,e):
        logging.debug("ConnectorDB: Websocket error: %s",str(e))
        v = self.unlockopen()
        if not v or self.isretry:
            if not self.isconnected and self.wantsconnection:
                self.isretry = True
                logging.warn("Disconnected from websocket. Retrying in %.2fs"%(self.reconnectbackoff,))
                #The connection was already unlocked, is not connected, and it wants a connection.
                reconnector = threading.Timer(self.reconnectbackoff,self.__reconnect_callback)
                reconnector.daemon=True
                reconnector.start()
                
    def __reconnect_callback(self):
        #Random reconnect times will have the server not get pounded if disconnect happens 
        self.reconnectbackoff += random.uniform(-1,5)
        if self.reconnectbackoff < 1.0:
            self.reconnectbackoff=1.0
        elif self.reconnectbackoff > 10*60.0:
            #If more than 10 minutes, make next reconnect in 10 minutes += 1 minute
            self.reconnectbackoff = 10*60.0 + random.uniform(-60,60)
        try:
            logging.debug("Reconnecting websocket...")
            self.connect()
            self.__resubscribe()
            logging.warn("Reconnect Successful")
        except:
            pass
    def __resubscribe(self):
        #Subscribe to all existing subscriptions (happens on reconnect)
        self.subscription_lock.acquire()
        for sub in self.subscriptions:
            logging.debug("Resubscribing to %s",sub)
            self.send("subscribe",sub)
        self.subscription_lock.release()

    def send(self,cmd):
        self.ws_sendlock.acquire()
        try:
            self.ws.send(json.dumps(cmd))
            self.ws_sendlock.release()
        except:
            self.ws_sendlock.release()
            raise

    def insert(self,uri,data):
        if not self.connect():
            return False
        try:
            logging.debug("Inserting thru websocket")
            self.send({"cmd": "insert", "arg": uri,"d": data})
        except:
            return False
        return True

    def subscribe(self,uri,callback):
        if not self.connect():
            return False

        logging.debug("Subscribing to %s",uri)
        #Subscribes to the given uri with the given callback
        self.send({"cmd": "subscribe", "arg": uri})
        self.subscription_lock.acquire()
        self.subscriptions[uri] = callback
        self.subscription_lock.release()

        return True

    def unsubscribe(self,uri):
        logging.debug("Unsubscribing from %s",uri)
        try:
            self.send({"cmd": "unsubscribe", "arg": uri})
        except:
            pass
        #Unsubscribes from the given uri
        self.subscription_lock.acquire()
        del self.subscriptions[uri]
        self.subscription_lock.release()

    def disconnect(self):
        self.wantsconnection = False
        #Closes the connection if it exists
        if self.ws is not None:
            self.ws.close()

        self.subscription_lock.acquire()
        self.subscriptions = {}
        self.subscription_lock.release()

    def connect(self):
        if not self.isconnected and self.wantsconnection:
            return False    #Means that is in process of retrying
        self.wantsconnection = True
        #Connects to the server if there is no connection active
        if not self.isconnected:
            self.ws = websocket.WebSocketApp(self.uri,header=self.headers,
                                             on_message = self.__on_message,
                                             on_close = self.__on_close,
                                             on_open = self.__on_open,
                                             on_error = self.__on_error)

            self.ws_thread = threading.Thread(target=self.ws.run_forever)
            self.ws_thread.daemon=True

            self.ws_openlock.acquire()
            self.ws_thread.start()

            #The lock will be released once there is news from the connection
            #so we acquire and release it again
            self.ws_openlock.acquire()
            self.ws_openlock.release()

            if not self.isconnected:
                raise errors.ConnectionError("Could not connect to "+self.uri)
            else:
                self.isretry=False
        return True