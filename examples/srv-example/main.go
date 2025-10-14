package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/glebarez/sqlite"
	"github.com/gorilla/handlers"
	"gorm.io/gorm"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/encoding/xml"
	"github.com/dolanor/rip/providers/gormprovider"
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
	ro := []rip.EntityRouteOption{
		rip.WithCodecs(
			json.Codec,
			xml.Codec,
			html.NewEntityCodec("/users/"),
			html.NewEntityFormCodec("/users/"),
		),
		rip.WithMiddlewares(loggerMiddleware(logWriter)),
	}
	// end route option OMIT

	// start HandleFuncEntities OMIT

	type User struct {
		ID   string `rip:"id" db:"id,key"`
		Name string `db:"name"`
	}

	up := mapprovider.New[User](memLogger)

	http.HandleFunc(rip.HandleEntities("/users/", up, ro...))
	// end HandleFuncEntities OMIT

	for i := 0; i < 10; i++ {
		id := strconv.Itoa(i)
		up.Create(context.Background(), User{
			ID:   id,
			Name: "George-" + id,
		})
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		panic(err)
	}

	sqlLogger := slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))
	sup := gormprovider.New[User](db, sqlLogger)
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

	ro = append(ro, rip.WithCodecs(
		// overwrite codec with another path
		html.NewEntityCodec("/sqlusers/"),
		html.NewEntityFormCodec("/sqlusers/"),
	))

	http.HandleFunc(rip.HandleEntities("/sqlusers/", sup, ro...))

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
