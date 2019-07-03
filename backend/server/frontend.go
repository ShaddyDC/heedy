package server

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/spf13/afero"
)

type fContext struct {
	User        *database.User                        `json:"user"`
	Admin       bool                                  `json:"admin"`
	Routes      map[string]string                     `json:"routes"`
	Menu        map[string]assets.MenuItem            `json:"menu"`
	SourceTypes map[string]*assets.SourceTypeFrontend `json:"source_types"`
}

type aContext struct {
	User    *database.User `json:"user"`
	Request *CodeRequest   `json:"request"`
}

// FrontendMux represents the frontend
func FrontendMux() (*chi.Mux, error) {
	mux := chi.NewMux()

	frontendFS := afero.NewBasePathFs(assets.Get().FS, "/public")

	// The main frontend app

	fbytes, err := afero.ReadFile(frontendFS, "/index.html")
	if err != nil {
		return nil, err
	}
	fTemplate, err := template.New("frontend").Parse(string(fbytes))
	if err != nil {
		return nil, err
	}

	// This is the main function that sets up the frontend template
	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Disallow clickjacking
		w.Header().Add("X-Frame-Options", "DENY")
		w.Header().Add("Cache-Control", "private, no-cache")

		ctx := CTX(r)
		u, err := ctx.DB.ReadUser(ctx.DB.ID(), &database.ReadUserOptions{
			Avatar: true,
		})
		if err != nil {
			WriteJSONError(w, r, http.StatusInternalServerError, err)
			return
		}

		cfg := assets.Config()

		sourceMap := make(map[string]*assets.SourceTypeFrontend)
		for k, v := range cfg.SourceTypes {
			sourceMap[k] = v.Frontend
		}

		if u == nil {
			// Running template as public
			err = fTemplate.Execute(w, &fContext{
				User:        nil,
				Admin:       false,
				Routes:      cfg.Frontend.PublicRoutes,
				Menu:        cfg.Frontend.PublicMenu,
				SourceTypes: sourceMap,
			})
			if err != nil {
				WriteJSONError(w, r, http.StatusInternalServerError, err)
			}
			return
		}

		err = fTemplate.Execute(w, &fContext{
			User:        u,
			Admin:       ctx.DB.AdminDB().Assets().Config.UserIsAdmin(*u.Name),
			Routes:      assets.Config().Frontend.Routes,
			Menu:        assets.Config().Frontend.Menu,
			SourceTypes: sourceMap,
		})
		if err != nil {
			WriteJSONError(w, r, http.StatusInternalServerError, err)
		}
		return

	})

	// Handles getting all assets other than the root webpage
	mux.Mount("/static/", http.FileServer(afero.NewHttpFs(frontendFS)))

	return mux, nil
}
