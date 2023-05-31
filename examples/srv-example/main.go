package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dolanor/rip"
	"github.com/gorilla/handlers"
)

const (
	defaultPort = "8888"
)

func logHandler(w io.Writer) func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return handlers.LoggingHandler(w, f).ServeHTTP
	}
}

func main() {
	hostPort := os.ExpandEnv("$HOST:$PORT")
	if hostPort == ":" {
		hostPort += defaultPort
	}

	up := NewUserProvider()
	http.HandleFunc("/greet", rip.Handle(http.MethodPost, Greet))
	http.HandleFunc(rip.HandleResource[*User]("/users/", up, logHandler(os.Stdout)))
	http.HandleFunc("/", handleRoot)

	fmt.Println("listening on " + hostPort)
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}
