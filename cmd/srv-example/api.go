package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/dolanor/rip"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World\n"))
}

func Greet(ctx context.Context, name string) (string, error) {
	return "Hello " + name, nil
}

type User struct {
	Name      string    `json:"name" xml:"name"`
	BirthDate time.Time `json:"birth_date" xml:"birth_date"`
}

func (u User) Identity() string {
	return u.Name
}

func (u *User) SetID(s string) {
	u.Name = s
}

func SaveUser(ctx context.Context, u *User) (*User, error) {
	log.Printf("SaveUser: saving %+v", u)
	mem[u.Name] = *u
	return u, nil
}

func GetUser(ctx context.Context, id string) (*User, error) {
	log.Printf("GetUser: getting %+v", id)
	u, ok := mem[id]
	if !ok {
		return &User{}, rip.NotFoundError{Resource: "user"}
	}
	return &u, nil
}

var mem = map[string]User{}
