package main

import (
	"flag"
	"github.com/rzetelskik/allezon-analytics/internal/allezon-analytics"
	"k8s.io/klog/v2"
	"runtime"
)

func main() {
	var err error

	klog.InitFlags(flag.CommandLine)
	err = flag.Set("logtostderr", "true")
	if err != nil {
		panic(err)
	}
	flag.Parse()
	defer klog.Flush()

	runtime.GOMAXPROCS(runtime.NumCPU())

	server := allezon_analytics.NewHTTPServer(":8080")

	klog.Info("Starting web server...")
	klog.Fatal(server.ListenAndServe())
}
