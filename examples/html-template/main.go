package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/examples/html-template/templates"
	"github.com/dolanor/rip/examples/srv-example/memuser"
)

func main() {
	hostPort := ":8989"

	ro := rip.NewRouteOptions().
		WithCodecs(
			html.NewEntityCodec("/users/", html.WithTemplatesFS(templates.FS)),
			html.NewEntityFormCodec("/users/", html.WithTemplatesFS(templates.FS)),
		)

	memLogger := log.New(os.Stderr, "in-memory: ", log.LstdFlags)
	up := memuser.NewUserProvider(memLogger)
	http.HandleFunc(rip.HandleEntities("/users/", up, ro))

	fmt.Println("listening on " + hostPort)
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}
