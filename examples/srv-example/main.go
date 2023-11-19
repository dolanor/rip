package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/dolanor/rip"
	_ "github.com/dolanor/rip/encoding/html"
	_ "github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/examples/srv-example/memuser"
)

const (
	defaultPort = "8888"
)

func logHandler(w io.Writer) func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %d", r.Method, r.URL.Path, r.ContentLength)
			f(w, r)
		}
	}
}

func main() {
	hostPort := os.ExpandEnv("$HOST:$PORT")
	if hostPort == ":" {
		hostPort += defaultPort
	}

	up := memuser.NewUserProvider()
	http.HandleFunc(rip.HandleEntity("/users/", up, logHandler(os.Stdout)))

	fmt.Println("listening on " + hostPort)
	go browse(hostPort)
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}
