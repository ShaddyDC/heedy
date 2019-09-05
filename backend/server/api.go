package server

import (
	"fmt"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/buildinfo"
)

func ReadUser(w http.ResponseWriter, r *http.Request) {
	var o database.ReadUserOptions
	username := chi.URLParam(r, "username")
	err := queryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	u, err := CTX(r).DB.ReadUser(username, &o)
	WriteJSON(w, r, u, err)
}
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var u database.User

	if err := UnmarshalRequest(r, &u); err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	u.ID = chi.URLParam(r, "username")
	WriteResult(w, r, CTX(r).DB.UpdateUser(&u))
}

func ListSources(w http.ResponseWriter,r *http.Request) {
	var o database.ListSourcesOptions
	err := queryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	sl,err := CTX(r).DB.ListSources(&o)
	WriteJSON(w, r, sl, err)
}

func CreateSource(w http.ResponseWriter, r *http.Request) {
	var s database.Source
	err := UnmarshalRequest(r, &s)
	if err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	adb := CTX(r).DB

	sid, err := adb.CreateSource(&s)
	if err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	s2, err := adb.ReadSource(sid, nil)

	WriteJSON(w, r, s2, err)
}

func ReadSource(w http.ResponseWriter, r *http.Request) {
	var o database.ReadSourceOptions
	srcid := chi.URLParam(r, "sourceid")
	err := queryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	s, err := CTX(r).DB.ReadSource(srcid, &o)
	WriteJSON(w, r, s, err)
}

func UpdateSource(w http.ResponseWriter, r *http.Request) {
	var s database.Source

	if err := UnmarshalRequest(r, &s); err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	s.ID = chi.URLParam(r, "sourceid")
	WriteResult(w, r, CTX(r).DB.UpdateSource(&s))
}

func DeleteSource(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "sourceid")
	WriteResult(w, r, CTX(r).DB.DelSource(sid))
}

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	var c database.Connection
	if err := UnmarshalRequest(r, &c); err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	db := CTX(r).DB
	cid,_, err := db.CreateConnection(&c)
	if err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	c2, err := db.ReadConnection(cid,&database.ReadConnectionOptions{
		APIKey: true,
	})
	WriteJSON(w,r,c2,err)
}

func ReadConnection(w http.ResponseWriter, r *http.Request) {
	var o database.ReadConnectionOptions
	cid := chi.URLParam(r, "connectionid")
	err := queryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	s, err := CTX(r).DB.ReadConnection(cid, &o)
	WriteJSON(w, r, s, err)
}


func UpdateConnection(w http.ResponseWriter, r *http.Request) {
	var c database.Connection

	if err := UnmarshalRequest(r, &c); err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	c.ID = chi.URLParam(r, "connectionid")
	WriteResult(w, r, CTX(r).DB.UpdateConnection(&c))
}

func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	cid := chi.URLParam(r, "connectionid")
	WriteResult(w, r, CTX(r).DB.DelConnection(cid))
}


func ListConnections(w http.ResponseWriter,r *http.Request) {
	var o database.ListConnectionOptions
	err := queryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	cl,err := CTX(r).DB.ListConnections(&o)
	WriteJSON(w, r, cl, err)
}


func GetSourceScopes(w http.ResponseWriter, r *http.Request) {
	// TODO: figure out whether to require auth for this
	a := CTX(r).DB.AdminDB().Assets()
	stype := chi.URLParam(r, "sourcetype")
	scopes, err := a.Config.GetSourceScopes(stype)
	WriteJSON(w,r,scopes,err)
}

func GetConnectionScopes(w http.ResponseWriter, r *http.Request) {
	a := CTX(r).DB.AdminDB().Assets()
	// Now our job is to generate all of the scopes
	// TODO: language support
	// TODO: maybe require auth for this?

	var smap = map[string]string{
		"owner": "All available access to your user",
		"owner:read": "Read your user info",
		"owner:update": "Modify your user's info",
		"users": "All permissions for all users",
		"users:read": "Read all users that you can read",
		"users:update": "Modify info for all users you can modify",
		"sources": "All permissions for all sources of all types",
		"sources:read": "Read all sources belonging to you (of all types)",
		"sources:update": "Modify data of all sources belonging to you (of all types)",
		"sources:delete": "Delete any sources belonging to you (of all types)",
		"shared": "All permissions for sources shared with you (of all types)",
		"shared:read": "Read sources of all types that were shared with you",
		"self.sources": "Allows the connection to create and manage its own sources of all types",
	}

	// Generate the source type scopes
	for stype := range a.Config.SourceTypes {
		smap[fmt.Sprintf("sources.%s",stype)] = fmt.Sprintf("All permissions for sources of type '%s'",stype)
		smap[fmt.Sprintf("sources.%s:read",stype)] = fmt.Sprintf("Read access for your sources of type '%s'",stype)
		smap[fmt.Sprintf("sources.%s:delete",stype)] = fmt.Sprintf("Can delete your sources of type '%s'",stype)

		smap[fmt.Sprintf("shared.%s",stype)] = fmt.Sprintf("All permissions for sources of type '%s' that were shared with you",stype)
		smap[fmt.Sprintf("shared.%s:read",stype)] = fmt.Sprintf("Read access for your sources of type '%s' that were shared with you",stype)
		
		smap[fmt.Sprintf("self.sources.%s",stype)] = fmt.Sprintf("Allows the connection to create and manage its own sources of type '%s'",stype)
	
		// And now generate the per-type scopes
		stypemap := a.Config.SourceTypes[stype].Scopes
		if stypemap!=nil {
			for sscope := range *stypemap {
				smap[fmt.Sprintf("sources.%s:%s",stype,sscope)] = (*stypemap)[sscope]
				//smap[fmt.Sprintf("self.sources.%s:%s",stype,sscope)] = (*stypemap)[sscope]
				smap[fmt.Sprintf("shared.%s:%s",stype,sscope)] = (*stypemap)[sscope]
			}
		}
	}

	WriteJSON(w,r,smap,nil)
	
}

func GetVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(buildinfo.Version))
}

func APINotFound(w http.ResponseWriter, r *http.Request) {
	WriteJSONError(w, r, http.StatusNotFound, errors.New("not_found: The given endpoint is not available"))
}

// APIMux gives the REST API
func APIMux() (*chi.Mux, error) {

	v1mux := chi.NewMux()

	v1mux.Get("/user/{username}", ReadUser)
	v1mux.Patch("/user/{username}", UpdateUser)

	v1mux.Post("/source", CreateSource)
	v1mux.Get("/source",ListSources)
	v1mux.Get("/source/{sourceid}", ReadSource)
	v1mux.Patch("/source/{sourceid}", UpdateSource)
	v1mux.Delete("/source/{sourceid}", DeleteSource)

	v1mux.Post("/connection", CreateConnection)
	v1mux.Get("/connection", ListConnections)
	v1mux.Get("/connection/{connectionid}",ReadConnection)
	v1mux.Patch("/connection/{connectionid}",UpdateConnection)
	v1mux.Delete("/connection/{connectionid}",DeleteConnection)

	v1mux.Get("/meta/scopes/{sourcetype}",GetSourceScopes)
	v1mux.Get("/meta/scopes", GetConnectionScopes)
	v1mux.Get("/meta/version",GetVersion)

	apiMux := chi.NewMux()
	apiMux.NotFound(APINotFound)
	apiMux.Mount("/heedy/v1", v1mux)
	return apiMux, nil
}
