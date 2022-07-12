package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rzetelskik/allezon-analytics/internal/allezon-analytics/aerospike"
	"github.com/rzetelskik/allezon-analytics/internal/allezon-analytics/api"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
)

type server struct {
	upStore *aerospike.AerospikeStore[api.UserProfile]
}

func (s *server) UserTagsPostHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		klog.ErrorS(err, "can't read io")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ut := api.UserTag{}
	err = json.Unmarshal(payload, &ut)
	if err != nil {
		klog.ErrorS(err, "can't unmarshall data")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	def := api.UserProfile{
		Views: make([]api.UserTag, 0),
		Buys:  make([]api.UserTag, 0),
	}

	f := func(xs []api.UserTag) func(int) bool {
		return func(i int) bool {
			return xs[i].Time.Before(ut.Time)
		}
	}
	modify := func(up *api.UserProfile) error {
		switch ut.Action {
		case api.VIEW:
			up.Views = HeadSlice(InsertIntoSortedSlice(ut, up.Views, f), UserTagPerActionLimit)
		case api.BUY:
			up.Buys = HeadSlice(InsertIntoSortedSlice(ut, up.Buys, f), UserTagPerActionLimit)
		}

		return nil
	}
	err = s.upStore.RMWWithGenCheck(ut.Cookie, 3, &def, modify)
	if err != nil {
		klog.ErrorS(err, "can't update user profile", "cookie", ut.Cookie)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *server) UserProfilesPostHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	cookie := mux.Vars(r)["cookie"]

	values := r.URL.Query()
	if !values.Has("time_range") {
		http.Error(w, "required parameter 'time_range' is missing", http.StatusBadRequest)
		return
	}

	timeRange := values.Get("time_range")
	lowerBound, upperBound, err := api.ParseTimeRange(timeRange)
	if err != nil {
		http.Error(w, fmt.Errorf("can't parse time range: %w", err).Error(), http.StatusBadRequest)
		return
	}

	var limit int
	if values.Has("limit") {
		limit, err = strconv.Atoi(values.Get("limit"))
		if err != nil {
			http.Error(w, "optional parameter 'limit' is invalid", http.StatusBadRequest)
			return
		}
	} else {
		limit = 200
	}

	up := api.UserProfile{}
	err = s.upStore.Get(cookie, &up)
	if err != nil {
		klog.ErrorS(err, "can't get user profile", "cookie", cookie)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filterFunc := func(x api.UserTag) bool {
		return (x.Time.After(lowerBound) || x.Time.Equal(lowerBound)) && x.Time.Before(upperBound)
	}

	up.Buys = HeadSlice(FilterSlice(up.Buys, filterFunc), limit)
	up.Views = HeadSlice(FilterSlice(up.Views, filterFunc), limit)

	upr := api.UserProfileResponse{
		Cookie:      cookie,
		UserProfile: up,
	}

	payload, err := json.Marshal(upr)
	if err != nil {
		klog.ErrorS(err, "can't marshal data", "cookie", cookie)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

func (s *server) AggregatesPostHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		klog.Errorf("can't read io: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *server) HealthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *server) ReadyzHandler(w http.ResponseWriter, _ *http.Request) {
	var err error

	if !s.upStore.Client.IsConnected() {
		err = errors.New("readyz probe: can't connect with database")
		klog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func NewHTTPServer(addr string, userProfileStore *aerospike.AerospikeStore[api.UserProfile]) *http.Server {
	s := &server{
		upStore: userProfileStore,
	}

	r := mux.NewRouter()

	r.HandleFunc("/user_tags", s.UserTagsPostHandler).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	r.HandleFunc("/user_profiles/{cookie}", s.UserProfilesPostHandler).
		Methods(http.MethodPost)

	r.HandleFunc("/aggregates", s.AggregatesPostHandler).
		Methods(http.MethodPost)

	r.HandleFunc("/healthz", s.HealthzHandler).
		Methods(http.MethodGet)

	r.HandleFunc("/readyz", s.ReadyzHandler).
		Methods(http.MethodGet)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
