package server

import (
	"encoding/json"
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
	modify := func(up *api.UserProfile) error {
		switch ut.Action {
		case api.VIEW:
			up.Views = LimitSlice(InsertIntoSortedSlice(ut, up.Views, func(i int) bool {
				return up.Views[i].Time.Before(ut.Time)
			}), 200)
		case api.BUY:
			up.Buys = LimitSlice(InsertIntoSortedSlice(ut, up.Buys, func(i int) bool {
				return up.Buys[i].Time.Before(ut.Time)
			}), 200)
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

	up.Buys = LimitSlice(FilterSlice(up.Buys, filterFunc), limit)
	up.Views = LimitSlice(FilterSlice(up.Views, filterFunc), limit)

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

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
