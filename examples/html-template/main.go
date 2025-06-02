package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/examples/html-template/templates"
	"github.com/dolanor/rip/providers/mapprovider"
)

type User struct {
	ID    string
	Name  string
	Email string
}

func main() {
	hostPort := ":8888"

	ro := []rip.EntityRouteOption{
		rip.WithCodecs(
			html.NewEntityCodec("/users/", html.WithTemplatesFS(templates.FS)),
			html.NewEntityFormCodec("/users/", html.WithTemplatesFS(templates.FS)),
		),
	}

	memLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	up := mapprovider.New[User](memLogger)
	http.HandleFunc(rip.HandleEntities("/users/", up, ro...))

	fmt.Println("listening on " + hostPort)
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}
