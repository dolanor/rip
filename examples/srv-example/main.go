package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dolanor/rip"
)

func main() {
	log.SetPrefix("DBGTHE: APP: ")
	hostPort := os.ExpandEnv("$HOST:$PORT")

	up := NewUserProvider()
	http.HandleFunc("/greet", rip.Handle(http.MethodPost, Greet))
	http.HandleFunc(rip.HandleResource[*User, *UserProvider]("/users/", up))
	http.HandleFunc("/", handleRoot)

	fmt.Println("listening on " + hostPort)
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}
