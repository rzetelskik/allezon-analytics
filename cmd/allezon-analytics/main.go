package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"
)

type UserTag struct {
	Time string `json:"time"` // format: "2022-03-22T12:15:00.000Z"
	//   millisecond precision
	//   with 'Z' suffix
	Cookie      string `json:"cookie"`
	Country     string `json:"country"`
	Device      string `json:"device"` // FIXME: enum
	Action      string `json:"action"` // FIXME: enum
	Origin      string `json:"origin"`
	ProductInfo struct {
		ProductID  uint64 `json:"product_id"` // FIXME: it's supposed to be a string
		BrandID    string `json:"brand_id"`
		CategoryID string `json:"category_id"`
		Price      int32  `json:"price"`
	} `json:"product_info"`
}

func UserTagsPostHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		klog.Errorf("can't read io: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ut := UserTag{}
	err = json.Unmarshal(data, &ut)
	if err != nil {
		klog.Errorf("can't unmarshall data: %w", err)
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

func UserProfilesPostHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		klog.Errorf("can't read io: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func AggregatesPostHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		klog.Errorf("can't read io: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func main() {
	klog.InitFlags(flag.CommandLine)
	err := flag.Set("logtostderr", "true")
	if err != nil {
		panic(err)
	}
	flag.Parse()
	defer klog.Flush()

	r := mux.NewRouter()

	r.HandleFunc("/user_tags", UserTagsPostHandler).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	r.HandleFunc("/user_profiles/{cookie}", UserProfilesPostHandler).
		Methods(http.MethodPost)

	r.HandleFunc("/aggregates", AggregatesPostHandler).
		Methods(http.MethodPost)

	klog.Info("Starting web server...")
	klog.Fatal(http.ListenAndServe(":8080", r))
}
