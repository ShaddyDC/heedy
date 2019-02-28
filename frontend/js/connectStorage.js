/*
  The frontend stores users/devices/streams locally in the browser, so that the app can be used offline.
  We therefore need a method to simply access these values, without worrying about whether they are available,
  or in our local storage, or if we are querying them, or whatever may be happening.

  While storage.js implements the actual querying, connectStorage is a special wrapper element, which allows any
  child component to go from user/device/stream names, to having actual component data.

  For example, connectStorage will go from the STRINGS 'myuser' 'mydevice' 'mystream', and give to its child component
  the actual myuser,mydevice, and mystream OBJECTS which contain the actual data.

  Furthermore, connectStorage is subscribed to the storage itself, so whenever there are updates using the REST api,
  the values are immediately reflected in the outputs, so all child components are always up to date.

  Usage:
    connectStorage(myComponent,false,false);
*/

// TODO: I cry when I see code like this. What makes it all the more horrible is that *I* am
//  the person who wrote it... This really needs to be refactored... - dkumor

import React, { Component } from "react";
import PropTypes from "prop-types";
import storage from "./storage";
import createClass from "create-react-class";

const NoQueryIfWithinMilliseconds = 1000;

export default function connectStorage(Component, lsdev, lsstream) {
  return createClass({
    propTypes: {
      user: PropTypes.string,
      device: PropTypes.string,
      stream: PropTypes.string,
      match: PropTypes.shape({
        params: PropTypes.shape({
          user: PropTypes.string,
          device: PropTypes.string,
          stream: PropTypes.string
        })
      })
    },
    getUser(props) {
      if (props === undefined) {
        props = this.props;
      }
      if (props.user !== undefined) return props.user;
      if (props.match.params.user !== undefined) return props.match.params.user;
      return "";
    },
    getDevice(props) {
      if (props === undefined) {
        props = this.props;
      }
      if (props.device !== undefined) return props.device;
      if (props.match.params.device !== undefined)
        return props.match.params.device;
      return "";
    },
    getStream(props) {
      if (props === undefined) {
        props = this.props;
      }
      if (props.stream !== undefined) return props.stream;
      if (props.match.params.stream !== undefined)
        return props.match.params.stream;
      return "";
    },
    getInitialState: function() {
      return {
        user: null,
        device: null,
        stream: null,
        error: null,
        devarray: null,
        streamarray: null
      };
    },
    getData: function(nextProps) {
      var thisUser = this.getUser(nextProps);
      // Get the user/device/stream from cache - this allows the app to feel fast in
      // slow internet, and enables working in offline mode
      storage.get(thisUser).then(response => {
        if (response != null) {
          if (response.ref !== undefined) {
            this.setState({ error: response });
          } else if (response.name !== undefined) {
            this.setState({ user: response });
            // If the user was recently queried, don't query it again needlessly
            if (response.timestamp > Date.now() - NoQueryIfWithinMilliseconds) {
              return;
            }
          }
        }
        // The query will be caught by the callback
        storage.query(thisUser).catch(err => console.log(err));
      });
      if (this.getDevice(nextProps) != "") {
        var thisDevice = thisUser + "/" + this.getDevice(nextProps);
        storage.get(thisDevice).then(response => {
          if (response != null) {
            if (response.ref !== undefined) {
              this.setState({ error: response });
            } else if (response.name !== undefined) {
              this.setState({ device: response });
              // If the user was recently queried, don't query it again needlessly
              if (
                response.timestamp >
                Date.now() - NoQueryIfWithinMilliseconds
              ) {
                return;
              }
            }
          }
          // The query will be caught by the callback
          storage.query(thisDevice).catch(err => console.log(err));
        });
        if (this.getStream(nextProps) != "") {
          var thisStream = thisDevice + "/" + this.getStream(nextProps);
          storage.get(thisStream).then(response => {
            if (response != null) {
              if (response.ref !== undefined) {
                this.setState({ error: response });
              } else if (response.name !== undefined) {
                this.setState({ stream: response });
                // If the user was recently queried, don't query it again needlessly
                if (
                  response.timestamp >
                  Date.now() - NoQueryIfWithinMilliseconds
                ) {
                  return;
                }
              }
            }
            // The query will be caught by the callback
            storage.query(thisStream).catch(err => console.log(err));
          });
        }
      }
      //Whether or not to add lists of children
      if (lsdev) {
        storage.ls(thisUser).then(response => {
          if (response.ref !== undefined) {
            this.setState({ error: response });
          } else {
            this.setState({ devarray: response });
          }
          // The query will be caught by the callback
          storage.query_ls(thisUser).catch(err => console.log(err));
        });
      }
      if (lsstream) {
        storage.ls(thisDevice).then(response => {
          if (response.ref !== undefined) {
            this.setState({ error: response });
          } else {
            this.setState({ streamarray: response });
          }
          // The query will be caught by the callback
          storage.query_ls(thisDevice); //.catch(err => console.log(err));
        });
      }
    },
    componentWillMount: function() {
      // https://stackoverflow.com/questions/1349404/generate-a-string-of-5-random-characters-in-javascript
      this.callbackID = Math.random().toString(36).substring(7);

      // Add the callback for storage
      storage.addCallback(this.callbackID, (path, obj) => {
        // If the current user/device/stream was updated, update the view
        let thisUser = this.getUser();
        let thisDevice = thisUser + "/" + this.getDevice();
        let thisStream = thisDevice + "/" + this.getStream();
        if (path == thisUser) {
          if (obj.ref !== undefined) {
            this.setState({ error: obj });
          } else {
            this.setState({ user: obj });
          }
        } else if (path == thisDevice) {
          if (obj.ref !== undefined) {
            this.setState({ error: obj });
          } else {
            this.setState({ device: obj });
          }
        } else if (path == thisStream) {
          if (obj.ref !== undefined) {
            this.setState({ error: obj });
          } else {
            this.setState({ stream: obj });
          }
        } else if (
          (lsdev || lsstream) &&
          obj.ref === undefined &&
          path.startsWith(thisUser + "/")
        ) {
          // We might want to update our arrays
          let p = path.split("/");
          switch (p.length) {
            case 2:
              if (lsdev) {
                let ndevarray = Object.assign({}, this.state.devarray);
                ndevarray[path] = obj;
                this.setState({ devarray: ndevarray });
              }
              break;
            case 3:
              if (p[1] == this.getDevice() && lsstream) {
                let nsarray = Object.assign({}, this.state.streamarray);
                nsarray[path] = obj;
                this.setState({ streamarray: nsarray });
              }
              break;
          }
        }
      });
      this.getData(this.props);
    },
    componentWillUnmount() {
      storage.remCallback(this.callbackID);
    },
    componentWillReceiveProps(nextProps) {
      if (
        this.getUser() != this.getUser(nextProps) ||
        this.getDevice() != this.getDevice(nextProps) ||
        this.getStream() != this.getStream(nextProps)
      ) {
        this.setState({
          user: null,
          device: null,
          stream: null,
          devarray: null,
          streamarray: null,
          error: null
        });

        this.getData(nextProps);
      }
    },
    render: function() {
      return (
        <Component
          {...this.props}
          user={this.state.user}
          device={this.state.device}
          stream={this.state.stream}
          error={this.state.error}
          devarray={this.state.devarray}
          streamarray={this.state.streamarray}
        />
      );
    }
  });
}
