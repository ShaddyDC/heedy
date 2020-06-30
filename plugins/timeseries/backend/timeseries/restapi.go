package timeseries

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/heedy/heedy/api/golang/plugin"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
)

var queryDecoder = schema.NewDecoder()

type TimeseriesInfo struct {
	plugin.ObjectInfo
	Schema map[string]interface{}
	Actor  bool
}

var ErrNotActor = errors.New("not_actor: The given timeseries does not accept actions")

func GetTimeseriesInfo(r *http.Request) (*TimeseriesInfo, error) {
	si, err := plugin.GetObjectInfo(r)
	if err != nil {
		return nil, err
	}
	schemaInterface, ok := si.Meta["schema"]
	if !ok {
		return nil, plugin.ErrPlugin("Timeseries metadata does not include schema")
	}
	schemaMap, ok := schemaInterface.(map[string]interface{})
	if !ok {
		return nil, plugin.ErrPlugin("Timeseries schema invalid")
	}
	actorInterface, ok := si.Meta["actor"]
	if !ok {
		return nil, plugin.ErrPlugin("Timeseries has incomplete metadata")
	}
	actor, ok := actorInterface.(bool)
	if !ok {
		return nil, plugin.ErrPlugin("Timeseries actor info invalid")
	}
	return &TimeseriesInfo{
		ObjectInfo: *si,
		Schema:     schemaMap,
		Actor:      actor,
	}, nil
}

func validateRequest(w http.ResponseWriter, r *http.Request, scope string) (*TimeseriesInfo, bool) {
	si, err := GetTimeseriesInfo(r)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return nil, false
	}
	if !si.ObjectInfo.Access.HasScope(scope) {
		rest.WriteJSONError(w, r, http.StatusForbidden, database.ErrAccessDenied("Insufficient permissions"))
		return nil, false
	}
	return si, true
}

func ReadData(w http.ResponseWriter, r *http.Request, action bool) {
	c := rest.CTX(r)
	si, ok := validateRequest(w, r, "read")
	if !ok {
		return
	}
	if action && !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	var q Query

	err := queryDecoder.Decode(&q, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	q.Actions = &action
	if q.Timeseries != "" {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("timeseries arg is set automatically when querying objects"))
		return
	}
	q.Timeseries = si.ObjectInfo.ID

	di, err := OpenSQLData(c.DB.AdminDB()).ReadTimeseriesData(&q)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	ai, err := NewJsonArrayReader(di)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	err = rest.WriteGZIP(w, r, ai, http.StatusOK)
	if err != nil {
		c.Log.Warnf("Timeseries read failed: %s", err.Error())
	}

}

func DeleteData(w http.ResponseWriter, r *http.Request, action bool) {
	c := rest.CTX(r)
	si, ok := validateRequest(w, r, "write")
	if !ok {
		return
	}
	if action && !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	var q Query

	err := queryDecoder.Decode(&q, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	q.Actions = &action
	if q.Timeseries != "" {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("timeseries arg is set automatically when querying objects"))
		return
	}
	q.Timeseries = si.ObjectInfo.ID

	err = OpenSQLData(c.DB.AdminDB()).RemoveTimeseriesData(&q)
	if err == nil {
		c.Events.Fire(&events.Event{
			Event:  "timeseries_data_delete",
			Object: si.ObjectInfo.ID,
			Data:   q,
		})
	}
	rest.WriteResult(w, r, err)
}

func shouldUpdateModifed(d *string) bool {
	if d == nil {
		return true
	}
	t, err := time.Parse("2006-01-02", *d)
	if err != nil {
		return true
	}
	cy, cm, cd := time.Now().UTC().Date()
	dy, dm, dd := t.Date()
	return cd > dd || cm > dm || cy > dy
}

type TimeseriesWriteEvent struct {
	T1    float64    `json:"t1"`
	T2    float64    `json:"t2"`
	Count int64      `json:"count"`
	DP    *Datapoint `json:"dp,omitempty"`
}

func WriteData(w http.ResponseWriter, r *http.Request, action bool) {
	c := rest.CTX(r)
	si, ok := validateRequest(w, r, "write")
	if !ok {
		return
	}
	if action && !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	var iq InsertQuery
	err := queryDecoder.Decode(&iq, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	iq.Actions = &action

	if action && !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}

	var datapoints DatapointArray

	err = rest.UnmarshalRequest(r, &datapoints)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	dv, err := NewDataValidator(NewDatapointArrayIterator(datapoints), si.Schema, c.DB.ID())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}

	dp, tstart, tend, count, err := OpenSQLData(c.DB.AdminDB()).WriteTimeseriesData(si.ObjectInfo.ID, dv, &iq)
	if err == nil && count > 0 {
		if shouldUpdateModifed(si.LastModified) {
			ne := database.Date(time.Now().UTC())
			// The timeseries is now non-empty, so label it as such
			err = c.DB.AdminDB().UpdateObject(&database.Object{
				Details: database.Details{
					ID: si.ID,
				},
				LastModified: &ne,
			})
		}
		evt := "timeseries_data_write"
		if action {
			evt = "timeseries_actions_write"
		}
		c.Events.Fire(&events.Event{
			Event:  evt,
			Object: si.ObjectInfo.ID,
			Data: &TimeseriesWriteEvent{
				T1:    tstart,
				T2:    tend,
				Count: count,
				DP:    dp,
			},
		})
	}

	rest.WriteResult(w, r, err)
}

func DataLength(w http.ResponseWriter, r *http.Request, action bool) {
	c := rest.CTX(r)
	si, ok := validateRequest(w, r, "read")
	if !ok {
		return
	}
	if action && !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	l, err := OpenSQLData(c.DB.AdminDB()).TimeseriesDataLength(si.ObjectInfo.ID, action)
	rest.WriteJSON(w, r, l, err)
}

// Act is given just the data portion of a datapoint, and it is inserted at the current timestamp
func Act(w http.ResponseWriter, r *http.Request) {
	c := rest.CTX(r)
	si, ok := validateRequest(w, r, "act")
	if !ok {
		return
	}
	if !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	var i interface{}
	err := rest.UnmarshalRequest(r, &i)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	dv, err := NewDataValidator(NewDatapointArrayIterator(DatapointArray{NewDatapoint(i)}), si.Schema, c.DB.ID())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	t := "append"
	a := true

	dp, tstart, tend, count, err := OpenSQLData(c.DB.AdminDB()).WriteTimeseriesData(si.ObjectInfo.ID, dv, &InsertQuery{
		Method:  &t,
		Actions: &a,
	})

	if err == nil && count > 0 {
		if shouldUpdateModifed(si.LastModified) {
			ne := database.Date(time.Now().UTC())
			// The timeseries is now non-empty, so label it as such
			err = c.DB.AdminDB().UpdateObject(&database.Object{
				Details: database.Details{
					ID: si.ID,
				},
				LastModified: &ne,
			})
		}
		c.Events.Fire(&events.Event{
			Event:  "timeseries_actions_write",
			Object: si.ObjectInfo.ID,
			Data: &TimeseriesWriteEvent{
				T1:    tstart,
				T2:    tend,
				Count: count,
				DP:    dp,
			},
		})
	}

	rest.WriteResult(w, r, err)
}

func GenerateDataset(w http.ResponseWriter, r *http.Request) {
	// Generate a dataset
	c := rest.CTX(r)
	var d Dataset
	err := rest.UnmarshalRequest(r, &d)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	di, err := d.Get(c.DB)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	pi := &FromPipeIterator{dpi: di, it: di}

	ai, err := NewJsonArrayReader(pi)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	err = rest.WriteGZIP(w, r, ai, http.StatusOK)
	if err != nil {
		c.Log.Warnf("Dataset read failed: %s", err.Error())
	}
}

// Handler is the global router for the timeseries API
var Handler = func() *chi.Mux {
	m := chi.NewMux()

	m.Get("/object/timeseries", func(w http.ResponseWriter, r *http.Request) {
		ReadData(w, r, false)
	})
	m.Delete("/object/timeseries", func(w http.ResponseWriter, r *http.Request) {
		DeleteData(w, r, false)
	})
	m.Post("/object/timeseries", func(w http.ResponseWriter, r *http.Request) {
		WriteData(w, r, false)
	})
	m.Get("/object/timeseries/length", func(w http.ResponseWriter, r *http.Request) {
		DataLength(w, r, false)
	})

	m.Get("/object/actions", func(w http.ResponseWriter, r *http.Request) {
		ReadData(w, r, true)
	})
	m.Delete("/object/actions", func(w http.ResponseWriter, r *http.Request) {
		DeleteData(w, r, true)
	})
	m.Post("/object/actions", func(w http.ResponseWriter, r *http.Request) {
		WriteData(w, r, true)
	})
	m.Get("/object/actions/length", func(w http.ResponseWriter, r *http.Request) {
		DataLength(w, r, true)
	})

	m.Post("/object/act", Act)

	m.Post("/api/dataset", GenerateDataset)

	m.NotFound(rest.NotFoundHandler)
	m.MethodNotAllowed(rest.NotFoundHandler)

	return m
}()
