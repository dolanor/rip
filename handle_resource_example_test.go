package rip_test

import (
	"context"
	"net/http"
	"time"

	"github.com/dolanor/rip"
)

func Example() {
	up := NewUserProvider()
	http.HandleFunc(rip.HandleResource("/users/", up))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

type User struct {
	Name      string    `json:"name" xml:"name"`
	BirthDate time.Time `json:"birth_date" xml:"birth_date"`
}

func (u User) IDString() string {
	return u.Name
}

func (u *User) IDFromString(s string) error {
	u.Name = s

	return nil
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
	up.mem[u.Name] = *u
	return u, nil
}

func (up UserProvider) Get(ctx context.Context, ider rip.IdentifiableResource) (*User, error) {
	u, ok := up.mem[ider.IDString()]
	if !ok {
		return &User{}, rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}
	return &u, nil
}

func (up *UserProvider) Delete(ctx context.Context, ider rip.IdentifiableResource) error {
	_, ok := up.mem[ider.IDString()]
	if !ok {
		return rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}

	delete(up.mem, ider.IDString())
	return nil
}

func (up *UserProvider) Update(ctx context.Context, u *User) error {
	_, ok := up.mem[u.Name]
	if !ok {
		return rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}
	up.mem[u.Name] = *u

	return nil
}

func (up *UserProvider) ListAll(ctx context.Context) ([]*User, error) {
	var users []*User
	for _, u := range up.mem {
		// we copy to avoid referring the same pointer that would get updated
		u := u
		users = append(users, &u)
	}

	return users, nil
}
