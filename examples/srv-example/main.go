package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	_ "modernc.org/sqlite"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/encoding/xml"
	"github.com/dolanor/rip/examples/srv-example/memuser"
	"github.com/dolanor/rip/examples/srv-example/sqluser"
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

	logWriter := &yellowWriter{w: os.Stderr}
	memLogger := log.New(logWriter, "inmem: ", log.LstdFlags)

	// start route option OMIT
	ro := rip.NewRouteOptions().
		WithCodecs(
			json.Codec,
			xml.Codec,
			html.Codec,
			html.FormCodec,
		).
		WithMiddlewares(loggerMW(logWriter))
	// end route option OMIT

	// start HandleFuncEntities OMIT
	up := memuser.NewUserProvider(memLogger)

	http.HandleFunc(rip.HandleEntities("/users/", up, ro))
	// end HandleFuncEntities OMIT

	db, err := sql.Open("sqlite", "users.db")
	if err != nil {
		panic(err)
	}

	sqlLogger := log.New(logWriter, "sql: ", log.LstdFlags)
	sup, err := sqluser.NewSQLUserProvider(db, sqlLogger)
	if err != nil {
		panic(err)
	}

	http.HandleFunc(rip.HandleEntities("/dbusers/", sup, ro))

	fmt.Println("listening on " + hostPort)
	go browse(hostPort)
	err = http.ListenAndServe(hostPort, nil)
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

func loggerMW(logOut io.Writer) func(http.HandlerFunc) http.HandlerFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		w := yellowWriter{w: logOut}
		logHandler := handlers.LoggingHandler(&w, hf)
		return logHandler.ServeHTTP
	}
}

type yellowWriter struct {
	w io.Writer
}

func (w *yellowWriter) Write(b []byte) (int, error) {
	w.w.Write([]byte{27, 91, 51, 51, 109})
	defer w.w.Write([]byte{27, 91, 48, 109})

	return w.w.Write(b)
}
