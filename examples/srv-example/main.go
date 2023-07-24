package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/examples/srv-example/memuser"
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

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World\n"))
}

func Greet(ctx context.Context, name string) (string, error) {
	return "Hello " + name, nil
}

func main() {
	hostPort := os.ExpandEnv("$HOST:$PORT")
	if hostPort == ":" {
		hostPort += defaultPort
	}

	up := memuser.NewUserProvider()
	http.HandleFunc(rip.HandleResource("/users/", up, logHandler(os.Stdout)))
	http.HandleFunc("/greet", rip.Handle(http.MethodPost, Greet))
	http.HandleFunc("/", handleRoot)

	fmt.Println("listening on " + hostPort)
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}
