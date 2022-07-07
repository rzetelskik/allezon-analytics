package allezon_analytics

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var views = MapStore{
	m: make(map[string][]UserTag),
}

var buys = MapStore{
	m: make(map[string][]UserTag),
}

const (
	UTC      = "2006-01-02T15:04:05"
	UTCMilli = "2006-01-02T15:04:05.000"
)

func UserTagsPostHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		klog.Errorf("can't read io: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ut := UserTag{}
	err = json.Unmarshal(data, &ut)
	if err != nil {
		klog.Errorf("can't unmarshall data: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var m *MapStore
	switch ut.Action {
	case VIEW:
		m = &views
	case BUY:
		m = &buys
	}

	err = m.Append(ut.Cookie, ut)
	if err != nil {
		klog.Errorf("can't append data: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UserProfileResponse struct {
	Cookie string    `json:"cookie"`
	Views  []UserTag `json:"views"`
	Buys   []UserTag `json:"buys"`
}

func ParseDateTimeSeconds(s string) (time.Time, error) {
	return time.Parse(UTC, s)
}

func ParseDatetimeMilliseconds(s string) (time.Time, error) {
	return time.Parse(UTCMilli, s)
}

func ParseTimeRange(s string, datetimeParseFunc func(string) (time.Time, error)) (time.Time, time.Time, error) {
	var err error

	trs := strings.Split(s, "_")
	if len(trs) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time range format")
	}

	lowerBound, err := datetimeParseFunc(trs[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid lower bound")
	}

	upperBound, err := datetimeParseFunc(trs[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid upper bound")
	}

	return lowerBound, upperBound, nil
}

func UserProfilesPostHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	cookie := mux.Vars(r)["cookie"]

	values := r.URL.Query()

	if !values.Has("time_range") {
		http.Error(w, "required parameter 'time_range' is missing", http.StatusBadRequest)
		return
	}

	timeRange := values.Get("time_range")
	lowerBound, upperBound, err := ParseTimeRange(timeRange, ParseDatetimeMilliseconds)
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

	/* GET VIEWS */
	vs, err := views.Get(cookie)
	if err != nil {
		klog.Errorf("can't get views for cookie %s: %v", cookie, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Filter
	vsTmp := make([]UserTag, 0)
	for _, v := range vs {
		if (v.Time.After(lowerBound) || v.Time.Equal(lowerBound)) && v.Time.Before(upperBound) {
			vsTmp = append(vsTmp, v)
		}
	}
	vs = vsTmp

	// Sort
	sort.Slice(vs, func(i, j int) bool {
		return vs[i].Time.After(vs[j].Time)
	})

	// Limit
	vs = vs[max(0, len(vs)-limit):]

	/* GET BUYS */
	bs, err := buys.Get(cookie)
	if err != nil {
		klog.Errorf("can't get buys for cookie %s: %v", cookie, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter
	bsTmp := make([]UserTag, 0)
	for _, b := range bs {
		if (b.Time.After(lowerBound) || b.Time.Equal(lowerBound)) && b.Time.Before(upperBound) {
			bsTmp = append(bsTmp, b)
		}
	}
	bs = bsTmp

	// Sort
	sort.Slice(bs, func(i, j int) bool {
		return bs[i].Time.After(bs[j].Time)
	})

	// Limit
	bs = bs[max(0, len(bs)-limit):]

	res := UserProfileResponse{
		Cookie: cookie,
		Views:  vs,
		Buys:   bs,
	}

	myData, err := json.Marshal(res)
	if err != nil {
		klog.Errorf("can't marshal data: %v", err)
	}

	/* REMOVE */
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		klog.Errorf("can't read io: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	upr := UserProfileResponse{}
	err = json.Unmarshal(data, &upr)
	if err != nil {
		klog.Errorf("can't unmarshall data: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !reflect.DeepEqual(upr, res) {
		klog.Errorf("cookie: %s\ntime_range: %s\nlimit: %d\nexpected and created data differ: %s", cookie, timeRange, limit, cmp.Diff(upr, res))
	}
	/* END REMOVE */

	w.Header().Set("Content-Type", "application/json")
	w.Write(myData)
}

func AggregatesPostHandler(w http.ResponseWriter, r *http.Request) {
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

func NewHTTPServer(addr string) *http.Server {
	r := mux.NewRouter()

	r.HandleFunc("/user_tags", UserTagsPostHandler).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	r.HandleFunc("/user_profiles/{cookie}", UserProfilesPostHandler).
		Methods(http.MethodPost)

	r.HandleFunc("/aggregates", AggregatesPostHandler).
		Methods(http.MethodPost)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
