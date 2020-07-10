package main

import (
	"net/http"
	"time"

	expvar2 "learn/copy/expvar/expvar"
)

//func standExpvar() {
//	lastAccess := expvar.NewString("lastAccess")
//	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
//		w.Write([]byte("hello"))
//		lastAccess.Set(time.Now().String())
//	})
//
//	http.ListenAndServe(":8080", nil)
//}

func customExpvar() {
	lastAccess := expvar2.NewString("lastAccess")
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
		lastAccess.Set(time.Now().String())
	})

	http.ListenAndServe(":8080", nil)
}

func main() {
	customExpvar()
}
