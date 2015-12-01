package website

import (
	"connectordb"
	"errors"
	"net/http"
	"server/webcore"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
)

//Authenticator runs an auth check and either goes to the www template given or to the apifunc handler
func Authenticator(www *FileTemplate, apifunc webcore.APIHandler, db *connectordb.Database) http.HandlerFunc {
	funcname := webcore.GetFuncName(apifunc)
	qtimer := webcore.GetQueryTimer(funcname)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		usr, ok := mux.Vars(request)["user"]
		if ok && usr == "api" || usr == WWWPrefix || usr == AppPrefix {
			//We don't want to pass 404s to the handlers
			NotFoundHandler(writer, request)
			return
		}

		tstart := time.Now()
		logger := webcore.GetRequestLogger(request, funcname)

		//We don't need the "op" here
		delete(logger.Data, "op")

		//Handle a panic without crashing the whole rest interface
		defer webcore.HandlePanic(logger)

		//Access control to the website is blocked
		//webcore.WriteAccessControlHeaders(writer, request)
		if !webcore.IsActive && webcore.HasSession(request) {
			WriteError(logger, writer, http.StatusServiceUnavailable, errors.New("ConnectorDB app is disabled for maintenance."), false)
			return
		}

		atomic.AddInt32(&webcore.StatsActive, 1)
		defer atomic.AddInt32(&webcore.StatsActive, -1)

		o, err := webcore.Authenticate(db, request)
		if err == nil {
			//There is a user logged in
			l := logger.WithField("dev", o.Name())

			//Only count valid web queries
			atomic.AddUint32(&webcore.StatsWebQueries, 1)

			loglevel, txt := apifunc(o, writer, request, l)

			//Find the time that this query took
			tdiff := time.Since(tstart)
			qtimer.Add(tdiff)

			webcore.LogRequest(l, loglevel, txt, tdiff)
			return
		}

		//If we got here, the user is not logged in. We therefore execute the "www" template given
		www.Execute(writer, map[string]string{
			"version": connectordb.Version,
		})

		webcore.LogRequest(logger, webcore.DEBUG, "", time.Since(tstart))
	})
}
