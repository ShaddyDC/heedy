package meta

import (
	"connectordb"
	"connectordb/query/interpolators"
	"connectordb/query/transforms"
	"net/http"
	"server/restapi/restcore"
	"server/webcore"
	"strconv"

	"github.com/gorilla/mux"
)

//TransformList returns the list of avaliable transforms and their descriptions
func TransformList(writer http.ResponseWriter, request *http.Request) {
	l := webcore.GetRequestLogger(request, "TransformList")

	webcore.WriteAccessControlHeaders(writer,request)
	restcore.JSONWriter(writer, transforms.Registry, l, nil)

}

//InterpolatorList returns the list of avaliable interpolators and their descriptions
func InterpolatorList(writer http.ResponseWriter, request *http.Request) {
	l := webcore.GetRequestLogger(request, "InterpolatorList")

	webcore.WriteAccessControlHeaders(writer,request)
	restcore.JSONWriter(writer, interpolators.Registry, l, nil)

}

//Version returns the ConnectorDB version being run
func Version(writer http.ResponseWriter, request *http.Request) {
	webcore.GetRequestLogger(request, "Version")

	webcore.WriteAccessControlHeaders(writer,request)
	writer.Header().Set("Content-Length", strconv.Itoa(len(connectordb.Version)))
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(connectordb.Version))
}

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *connectordb.Database, prefix *mux.Router) *mux.Router {
	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	prefix.HandleFunc("/transforms", http.HandlerFunc(TransformList)).Methods("GET")
	prefix.HandleFunc("/interpolators", http.HandlerFunc(InterpolatorList)).Methods("GET")
	prefix.HandleFunc("/version", http.HandlerFunc(Version)).Methods("GET")

	return prefix
}
