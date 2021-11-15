package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dolanor/rip"
)

func main() {
	hostPort := os.ExpandEnv("$HOST:$PORT")
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/greet", rip.Handle(http.MethodPost, Greet))
	http.HandleFunc("/users/", rip.HandleResource(User{}, SaveUser, GetUser))

	fmt.Println("listening on " + hostPort)
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}
