package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
	_ "modernc.org/sqlite"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/encoding/xml"
	"github.com/dolanor/rip/examples/srv-example/sqluser"
	"github.com/dolanor/rip/providers/mapprovider"
)

const (
	defaultPort = "8888"
)

func main() {
	hostPort := os.ExpandEnv("$HOST:$PORT")
	if hostPort == ":" {
		hostPort += defaultPort
	}

	logWriter := &yellowWriter{w: os.Stderr}
	memLogger := slog.New(slog.NewTextHandler(logWriter, nil))

	// start route option OMIT
	ro := rip.NewRouteOptions().
		WithCodecs(
			json.Codec,
			xml.Codec,
			html.NewEntityCodec("/users/"),
			html.NewEntityFormCodec("/users/"),
		).
		WithMiddlewares(loggerMiddleware(logWriter))
	// end route option OMIT

	// start HandleFuncEntities OMIT

	type User struct {
		Lol  string `rip:"id"`
		Name string
	}

	up := mapprovider.New[User](memLogger)
	for i := 0; i < 100; i++ {
		up.Create(context.Background(), User{
			Lol:  strconv.Itoa(i),
			Name: "Mohamed",
		})
	}

	http.HandleFunc(rip.HandleEntities("/users/", &up, ro))
	// end HandleFuncEntities OMIT

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}

	sqlLogger := log.New(logWriter, "sql: ", log.LstdFlags)
	sup, err := sqluser.NewSQLUserProvider(db, sqlLogger)
	if err != nil {
		panic(err)
	}

	ro = ro.
		WithCodecs(
			// overwrite codec with another path
			html.NewEntityCodec("/sqlusers/"),
			html.NewEntityFormCodec("/sqlusers/"),
		)

	http.HandleFunc(rip.HandleEntities("/sqlusers/", sup, ro))

	fmt.Println("listening on " + hostPort)
	//	go browse(hostPort)
	err = http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}

func loggerMiddleware(logOut io.Writer) func(http.HandlerFunc) http.HandlerFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		logHandler := handlers.LoggingHandler(logOut, hf)
		return logHandler.ServeHTTP
	}
}

// yellowWriter is just a io.Writer that writes in ANSI escape code
// to write in yellow.
type yellowWriter struct {
	w io.Writer
}

func (w *yellowWriter) Write(b []byte) (int, error) {
	w.w.Write([]byte{27, 91, 51, 51, 109})
	defer w.w.Write([]byte{27, 91, 48, 109})

	return w.w.Write(b)
}
