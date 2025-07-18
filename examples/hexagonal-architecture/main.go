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

	opts := []rip.EntityRouteOption{
		rip.WithCodecs(
			json.Codec,
			html.NewEntityCodec("/users/"),
			html.NewEntityFormCodec("/users/"),
		),
		rip.WithMiddlewares(loggerMiddleware(os.Stdout)),
		rip.WithErrors(rip.StatusMap{
			domain.ErrAppNotFound:       http.StatusNotFound,
			domain.ErrAppNotImplemented: http.StatusNotImplemented,
		}),
		rip.WithListPage(3, 10),
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	repo, err := sqladapter.NewUserRepo(db)
	if err != nil {
		log.Fatal(err)
	}

	up, err := domain.NewUserProvider(repo)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc(rip.HandleEntities("/users/", up, opts...))

	fmt.Println("listening on " + hostPort)
	err = http.ListenAndServe(hostPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
