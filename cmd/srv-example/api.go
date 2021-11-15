package main

import (
	"context"
	"net/http"

	"github.com/dolanor/rip"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World\n"))
}

func Greet(ctx context.Context, name string) (string, error) {
	return "Hello " + name, nil
}

type User struct {
	Name string
	Age  int
}

func SaveUser(ctx context.Context, u User) (User, error) {
	mem[u.Name] = u
	return u, nil
}

func GetUser(ctx context.Context, name string) (User, error) {
	u, ok := mem[name]
	if !ok {
		return User{}, rip.NotFoundError{Resource: "user"}
	}
	return u, nil
}

var mem = map[string]User{}
