package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/lovoo/goka"
	"github.com/rzetelskik/allezon-analytics/service/internal/service/aerospike"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type server struct {
	upStore *aerospike.AerospikeStore[api.UserProfile]
	emitter *goka.Emitter
	view    *goka.View
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

	err = s.emitter.EmitSync(ut.Cookie, payload)
	if err != nil {
		klog.ErrorS(err, "can't emit to kafka")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

type UserAggregates struct {
	Count    int64 `json:"count"`
	SumPrice int64 `json:"sum_price"`
}

func (s *server) AggregatesPostHandler(w http.ResponseWriter, r *http.Request) {
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

	// FIXME: check for max time range

	if !values.Has("action") {
		http.Error(w, "required parameter 'action' is missing", http.StatusBadRequest)
		return
	}
	action := values.Get("action")

	// FIXME: validate action

	if !values.Has("aggregates") {
		http.Error(w, "required parameter 'action' is missing", http.StatusBadRequest)
		return
	}

	aggregates := values["aggregates"]
	// FIXME: validate aggregates

	origin := values.Get("origin")           // FIXME
	brand_id := values.Get("brand_id")       // FIXME
	category_id := values.Get("category_id") // FIXME

	expected, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		klog.Errorf("can't read io: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ar := api.AggregatesResponse{
		Columns: []string{"1m_bucket", "action"},
		Rows:    [][]string{},
	}
	if len(origin) > 0 {
		ar.Columns = append(ar.Columns, "origin")
	}
	if len(brand_id) > 0 {
		ar.Columns = append(ar.Columns, "brand_id")
	}
	if len(category_id) > 0 {
		ar.Columns = append(ar.Columns, "category_id")
	}
	for _, a := range aggregates {
		ar.Columns = append(ar.Columns, strings.ToLower(a))
	}

	for b := lowerBound; b.Before(upperBound); b = b.Add(time.Minute) {
		key := b.String() + action + origin + brand_id + category_id
		klog.InfoS("key", key)
		h := sha256.New()
		h.Write([]byte(key))
		hash := hex.EncodeToString(h.Sum(nil))

		v, err := s.view.Get(hash)
		if err != nil {
			// FIXME: err aggregate
			// FIXME: check for error type and break
			klog.ErrorS(err, "tmp error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ua := UserAggregates{
			Count:    0,
			SumPrice: 0,
		}
		if v != nil {
			klog.InfoS("hurray, not nil!")
			ua = v.(UserAggregates)
		}

		row := []string{b.Format("2006-01-02T15:04:05"), action}
		if len(origin) > 0 {
			row = append(row, origin)
		}
		if len(brand_id) > 0 {
			row = append(row, brand_id)
		}
		if len(category_id) > 0 {
			row = append(row, category_id)
		}
		for _, a := range aggregates {
			var i int64
			switch a {
			case "COUNT":
				i = ua.Count
			case "SUM_PRICE":
				i = ua.SumPrice
			}
			row = append(row, strconv.FormatInt(i, 10))
		}

		ar.Rows = append(ar.Rows, row)
	}
	klog.InfoS("created aggregate", "aggregate", ar)

	data, err := json.Marshal(ar)
	if err != nil {
		klog.Errorf("can't marshall aggregate response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !reflect.DeepEqual(expected, data) {
		klog.Errorf("expected and actual data differ: %s", cmp.Diff(expected, data))
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

func NewHTTPServer(addr string, userProfileStore *aerospike.AerospikeStore[api.UserProfile], emitter *goka.Emitter, view *goka.View) *http.Server {
	s := &server{
		upStore: userProfileStore,
		emitter: emitter,
		view:    view,
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
