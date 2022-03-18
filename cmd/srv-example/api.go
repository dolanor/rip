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

func (u User) IDString() string {
	return u.Name
}

func (u *User) FromString(s string) {
	u.Name = s
}

type UserProvider struct {
	mem map[string]User
}

func NewUserProvider() *UserProvider {
	return &UserProvider{
		mem: map[string]User{},
	}
}

func (up *UserProvider) Create(ctx context.Context, u *User) (*User, error) {
	log.Printf("SaveUser: %+v", *u)
	up.mem[u.Name] = *u
	return u, nil
}

func (up UserProvider) Get(ctx context.Context, ider rip.IDer) (*User, error) {
	log.Printf("GetUser: %+v", ider.IDString())
	u, ok := up.mem[ider.IDString()]
	if !ok {
		return &User{}, rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}
	return &u, nil
}

func (up *UserProvider) Delete(ctx context.Context, ider rip.IDer) error {
	log.Printf("DeleteUser: %+v", ider.IDString())
	_, ok := up.mem[ider.IDString()]
	if !ok {
		return rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}

	delete(up.mem, ider.IDString())
	return nil
}

func (up *UserProvider) Update(ctx context.Context, u *User) error {
	log.Printf("UpdateUser: %+v", u.IDString())
	_, ok := up.mem[u.Name]
	if !ok {
		return rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}
	up.mem[u.Name] = *u

	return nil
}
