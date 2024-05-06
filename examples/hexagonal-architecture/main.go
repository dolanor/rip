package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	_ "modernc.org/sqlite"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/examples/hexagonal-architecture/adapters/sqladapter"
	"github.com/dolanor/rip/examples/hexagonal-architecture/domain"
)

func loggerMiddleware(logOut io.Writer) func(http.HandlerFunc) http.HandlerFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		logHandler := handlers.LoggingHandler(logOut, hf)
		return logHandler.ServeHTTP
	}
}

func main() {
	const (
		defaultPort = "8888"
	)

	hostPort := os.ExpandEnv("$HOST:$PORT")
	if hostPort == ":" {
		hostPort += defaultPort
	}

	ro := rip.NewRouteOptions().
		WithCodecs(
			json.Codec,
			html.NewEntityCodec("/users/"),
			html.NewEntityFormCodec("/users/"),
		).
		WithMiddlewares(loggerMiddleware(os.Stdout)).
		WithErrors(rip.ErrorMap{
			NotFound:          domain.ErrAppNotFound,
			ErrNotImplemented: domain.ErrAppNotImplemented,
		})

	db, err := sql.Open("sqlite", "local.db")
	if err != nil {
		log.Fatal(err)
	}

	repo, err := sqladapter.NewUserRepo(db)
	if err != nil {
		log.Fatal(err)
	}

	up, err := domain.NewSQLUserProvider(repo)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc(rip.HandleEntities("/users/", up, ro))

	fmt.Println("listening on " + hostPort)
	err = http.ListenAndServe(hostPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
