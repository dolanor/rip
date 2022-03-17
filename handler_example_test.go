package rip_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dolanor/rip"
)

func Example_handleResourceWithPath() {

	hostPort := os.ExpandEnv("$HOST:$PORT")

	up := NewUserProvider()
	http.HandleFunc("/greet", rip.Handle(http.MethodPost, Greet))
	http.HandleFunc(rip.HandleResourceWithPath[*User, *UserProvider]("/users/", up))

	fmt.Println("listening on " + hostPort)
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
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
		return &User{}, rip.NotFoundError{Resource: "user"}
	}
	return &u, nil
}

func (up *UserProvider) Delete(ctx context.Context, ider rip.IDer) error {
	log.Printf("DeleteUser: %+v", ider.IDString())
	_, ok := up.mem[ider.IDString()]
	if !ok {
		return rip.NotFoundError{Resource: "user"}
	}

	delete(up.mem, ider.IDString())
	return nil
}

func (up *UserProvider) Update(ctx context.Context, u *User) error {
	log.Printf("UpdateUser: %+v", u.IDString())
	_, ok := up.mem[u.Name]
	if !ok {
		return rip.NotFoundError{Resource: "user"}
	}
	up.mem[u.Name] = *u

	return nil
}
