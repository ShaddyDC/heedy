
import Vue from "../../dist.mjs";
import api from "../api.mjs";

let users = {};
if (appinfo.user != null) {
  users[appinfo.user.name] = appinfo.user;
}

export default {
    state: {
        alert: {
            value: false,
            text: "",
            type: ""
        },
        users: users,
        sources: {},

        // The current user's connections
        connections: null,

        // A list of IDs under each user's key
        userSources: {},

        // The following are initialized by the sourceInjector
        sourceCreators: [],

        // Subpaths for each source type
        typePaths: {},

        // The map of connection scopes along with their descriptions
        connectionScopes: null
    },
    mutations: {
        alert(state, v) {
          state.alert = {
            value: true,
            type: "",
            text: "",
            ...v
          };
        },
        setUser(state, v) {
          Vue.set(state.users, v.username, v);
        },
        setSource(state, v) {
          Vue.set(state.sources,v.id, v);
        },
        setConnection(state,v) {
          if (state.connections == null) {
            state.connections = {};
          }
          Vue.set(state.connections,v.id, v);
        },
        setConnections(state,v) {
          state.connections = v;
        },
        setUserSources(state,v) {
          let srcidarray = [];
          for (let i=0;i < v.sources.length;i++) {
            Vue.set(state.sources,v.sources[i].id, v.sources[i]);
            srcidarray.push(v.sources[i].id);
          }
          Vue.set(state.userSources,v.user,srcidarray);
        },
        setConnectionScopes(state,v) {
          state.connectionScopes = v;
        }
      },
      actions: {
        errnotify({ commit }, v) {
          // Notifies of an error
          if (v.hasOwnProperty("error")) {
            // Only notify if it is an actual error
            commit("alert", {
              type: "error",
              text: v.error_description
            });
          }
        },
        readUser: async function({ commit, rootState }, q) {
          console.log("Reading user", q.username);
          let res = await api("GET", `api/heedy/v1/users/${q.username}`, {
            avatar: true
          });
          if (!res.response.ok) {
            commit("alert", {
              type: "error",
              text: res.data.error_description
            });
            
          } else {
            if (rootState.app.info.user!=null && rootState.app.info.user.username == q.username) {
              commit("updateLoggedInUser",res.data);
            }
            commit("setUser", res.data);
          }
          
          
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
      },
      readSource: async function({ commit, rootState }, q) {
        console.log("Reading source", q.id);
        let res = await api("GET", `api/heedy/v1/sources/${q.id}`, {
          avatar: true
        });
        if (!res.response.ok) {
          commit("alert", {
            type: "error",
            text: res.data.error_description
          });
          
        } else {
          commit("setSource", res.data);
        }
        
        
        if (q.hasOwnProperty("callback")) {
          q.callback();
        }
      },
      readUserSources: async function({commit,rootState}, q) {
        console.log("Reading sources for user", q.username);
        let query = {username: q.username};

        if (rootState.app.info.user!=null && rootState.app.info.user.username == q.username) {
          query["connection"] = "none";
        }

        let res = await api("GET", `api/heedy/v1/sources`, query);
        if (!res.response.ok) {
          commit("alert", {
            type: "error",
            text: res.data.error_description
          });
          
        } else {
          commit("setUserSources", {user: q.name, sources: res.data});
        }
        
        
        if (q.hasOwnProperty("callback")) {
          q.callback();
        }

      },
      getConnectionScopes: async function({commit}) {
        console.log("Loading available connection scopes");
        let res = await api("GET", "api/heedy/v1/meta/scopes");
        if (!res.response.ok) {
          commit("alert", {
            type: "error",
            text: res.data.error_description
          });
          
        } else {
          commit("setConnectionScopes", res.data);
        }
      },
      listConnections: async function({commit}) {
        console.log("Loading connections");
        let res = await api("GET", "api/heedy/v1/connections",{avatar: true});
        if (!res.response.ok) {
          commit("alert", {
            type: "error",
            text: res.data.error_description
          });
          return;
        }
        let cmap = {};
        res.data.map(v => { cmap[v.id] = v });
        commit("setConnections",cmap);
      },
      readConnection: async function({ commit, rootState }, q) {
        console.log("Reading connection", q.id);
        let res = await api("GET", `api/heedy/v1/connections/${q.id}`, {
          avatar: true
        });
        if (!res.response.ok) {
          commit("alert", {
            type: "error",
            text: res.data.error_description
          });
          
        } else {
          commit("setConnection", res.data);
        }
        
        
        if (q.hasOwnProperty("callback")) {
          q.callback();
        }
      }

    }
    
};