package server

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/heedy/heedy/api/golang/rest"
)

/*
type appStruct {
	*database.App

	Unique bool `json:"unique"`
}

func GetPluginApps(w http.ResponseWriter, r *http.Request) {
	// Get all the apps available for creation
	a := rest.CTX(r).DB.AdminDB().Assets()

	db := rest.CTX(r).DB
	if db.Type() == database.PublicType || db.Type() == database.AppType {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only logged in users can list plugin apps"))
		return
	}



	m := make(map[string]appStruct)


}
*/

func APINotFound(w http.ResponseWriter, r *http.Request) {
	rest.WriteJSONError(w, r, http.StatusNotFound, rest.ErrNotFound)
}

// APIMux gives the REST API
func APIMux() (*chi.Mux, error) {

	v1mux := chi.NewMux()

	v1mux.Get("/events", EventWebsocket)
	v1mux.Post("/events", FireEvent)

	v1mux.Post("/users", CreateUser)
	v1mux.Get("/users", ListUsers)
	v1mux.Get("/users/{username}", ReadUser)
	v1mux.Patch("/users/{username}", UpdateUser)
	v1mux.Delete("/users/{username}", DeleteUser)

	v1mux.Post("/sources", CreateSource)
	v1mux.Get("/sources", ListSources)
	v1mux.Get("/sources/{sourceid}", ReadSource)
	v1mux.Patch("/sources/{sourceid}", UpdateSource)
	v1mux.Delete("/sources/{sourceid}", DeleteSource)

	v1mux.Post("/apps", CreateApp)
	v1mux.Get("/apps", ListApps)
	v1mux.Get("/apps/{appid}", ReadApp)
	v1mux.Patch("/apps/{appid}", UpdateApp)
	v1mux.Delete("/apps/{appid}", DeleteApp)

	v1mux.Get("/server/scopes/{sourcetype}", GetSourceScopes)
	v1mux.Get("/server/scopes", GetAppScopes)
	v1mux.Get("/server/apps", GetPluginApps)
	v1mux.Get("/server/version", GetVersion)

	v1mux.Get("/server/admin", GetAdminUsers)
	v1mux.Post("/server/admin/{username}", AddAdminUser)
	v1mux.Delete("/server/admin/{username}", RemoveAdminUser)

	v1mux.Get("/server/updates", GetUpdates)
	v1mux.Get("/server/updates/status", GetUpdateStatus)
	v1mux.Get("/server/updates/heedy.conf", GetConfigFile)
	v1mux.Post("/server/updates/heedy.conf", PostConfigFile)
	v1mux.Get("/server/updates/config", GetUConfig)
	v1mux.Patch("/server/updates/config", PatchUConfig)
	v1mux.Get("/server/updates/plugins", GetAllPlugins)
	v1mux.Post("/server/updates/plugins", PostPlugin)

	apiMux := chi.NewMux()
	apiMux.NotFound(APINotFound)
	apiMux.Mount("/heedy/v1", v1mux)
	return apiMux, nil
}
