package main

import (
	"expvar"
	"net/http"
	"time"
)

func main() {

	lastAccess := expvar.NewString("lastAccess")
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
		lastAccess.Set(time.Now().String())
	})

	http.ListenAndServe(":8080", nil)
}
