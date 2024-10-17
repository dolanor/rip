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
	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/sqlite"
	"github.com/samonzeweb/godb/tablenamer"
	_ "modernc.org/sqlite"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/encoding/xml"
	"github.com/dolanor/rip/providers/godbprovider"
	"github.com/dolanor/rip/providers/mapprovider"
)

const (
	defaultPort = "8888"
)

func main() {
	hostPort := os.ExpandEnv("$HOST:$PORT")
	if hostPort == ":" {
		hostPort = "localhost:" + defaultPort
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
		ID   string `rip:"id" db:"id,key"`
		Name string `db:"name"`
	}

	up := mapprovider.New[User](memLogger)
	for i := 0; i < 10; i++ {
		id := strconv.Itoa(i)
		up.Create(context.Background(), User{
			ID:   id,
			Name: "George-" + id,
		})
	}

	http.HandleFunc(rip.HandleEntities("/users/", up, ro))
	// end HandleFuncEntities OMIT

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users(id text primary key, name text);")
	if err != nil {
		panic(err)
	}

	godb := godb.Wrap(sqlite.Adapter, db)
	godb.SetLogger(log.New(os.Stderr, "", 0))
	godb.SetDefaultTableNamer(tablenamer.SnakePlural())

	sqlLogger := slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))
	sup := godbprovider.New[User](godb, sqlLogger)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		id := strconv.Itoa(i)
		sup.Create(context.Background(), User{
			ID:   id,
			Name: "George-" + id,
		})
	}

	ro = ro.
		WithCodecs(
			// overwrite codec with another path
			html.NewEntityCodec("/sqlusers/"),
			html.NewEntityFormCodec("/sqlusers/"),
		)

	http.HandleFunc(rip.HandleEntities("/sqlusers/", sup, ro))

	fmt.Println("check http://" + hostPort + "/users/ or http://" + hostPort + "/sqlusers/")
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
