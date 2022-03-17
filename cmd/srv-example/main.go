package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dolanor/rip"
)

func tuple() (string, http.HandlerFunc) {
	return "/lol", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "Hello World") }
}
func main() {
	hostPort := os.ExpandEnv("$HOST:$PORT")

	up := NewUserProvider()
	http.HandleFunc("/greet", rip.Handle(http.MethodPost, Greet))
	http.HandleFunc(rip.HandleResourcePath("/users/", &User{}, up.SaveUser, up.GetUser, up.UpdateUser, up.DeleteUser))
	http.HandleFunc(tuple())
	http.HandleFunc("/", handleRoot)

	fmt.Println("listening on " + hostPort)
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}
