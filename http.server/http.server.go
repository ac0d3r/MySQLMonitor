package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

var (
	flagHelp = flag.Bool("h", false, "Shows usage options.")
	flagHost = flag.String("host", "0.0.0.0", "listen host")
	flagPort = flag.Uint("port", 8080, "listen port")
	flagDir  = flag.String("dir", "./", "listen directory")
)

func getStatusCode(w http.ResponseWriter) int64 {
	respValue := reflect.ValueOf(w)
	if respValue.Kind() == reflect.Ptr {
		respValue = respValue.Elem()
	}
	status := respValue.FieldByName("status")
	return status.Int()
}

func withLog(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		log.Printf("handle %s %d\n", r.URL.Path, getStatusCode(w))
	})
}

func main() {
	var (
		addr string
	)
	flag.Parse()
	if *flagHelp {
		fmt.Printf("Usage: http.server [options]\n\n")
		flag.PrintDefaults()
		return
	}
	addr = fmt.Sprintf("%s:%d", *flagHost, *flagPort)
	log.Printf("listen on http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, withLog(http.FileServer(http.Dir(*flagDir)))))
}
