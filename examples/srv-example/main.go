package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dolanor/rip"
	_ "github.com/dolanor/rip/encoding/html"
	_ "github.com/dolanor/rip/encoding/json"
	_ "github.com/dolanor/rip/encoding/xml"
	"github.com/dolanor/rip/examples/srv-example/memuser"
)

const (
	defaultPort = "8888"
)

func uppercase(ctx context.Context, s string) (string, error) {
	u := strings.ToUpper(s)
	return u, nil
}

func main() {
	hostPort := os.ExpandEnv("$HOST:$PORT")
	if hostPort == ":" {
		hostPort += defaultPort
	}

	up := memuser.NewUserProvider()
	http.HandleFunc(rip.HandleEntities("/users/", up, logHandler(os.Stdout)))

	fmt.Println("listening on " + hostPort)
	go browse(hostPort)
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}

func logHandler(w io.Writer) func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %d", r.Method, r.URL.Path, r.ContentLength)
			f(w, r)
		}
	}
}
