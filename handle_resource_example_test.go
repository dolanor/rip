package rip

import (
	"context"
	"net/http"
	"time"

	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/encoding/json"
)

func Example() {
	up := newUserProvider()
	ro := NewRouteOptions().
		WithCodecs(json.Codec, html.NewEntityCodec("/users/"))
	http.HandleFunc(HandleEntities("/users/", up, ro))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

type user struct {
	Name         string    `json:"name" xml:"name"`
	EmailAddress string    `json:"email_address" xml:"email_address"`
	BirthDate    time.Time `json:"birth_date" xml:"birth_date"`
}

func (u user) IDString() string {
	return u.Name
}

func (u *user) IDFromString(s string) error {
	u.Name = s

	return nil
}

type UserProvider struct {
	mem map[string]user
}

func newUserProvider() *UserProvider {
	return &UserProvider{
		mem: map[string]user{},
	}
}

func (up *UserProvider) Create(ctx context.Context, u *user) (*user, error) {
	up.mem[u.Name] = *u
	return u, nil
}

func (up UserProvider) Get(ctx context.Context, entity Entity) (*user, error) {
	u, ok := up.mem[entity.IDString()]
	if !ok {
		return &user{}, ErrNotFound
	}
	return &u, nil
}

func (up *UserProvider) Delete(ctx context.Context, entity Entity) error {
	_, ok := up.mem[entity.IDString()]
	if !ok {
		return ErrNotFound
	}

	delete(up.mem, entity.IDString())
	return nil
}

func (up *UserProvider) Update(ctx context.Context, u *user) error {
	_, ok := up.mem[u.Name]
	if !ok {
		return ErrNotFound
	}
	up.mem[u.Name] = *u

	return nil
}

func (up *UserProvider) ListAll(ctx context.Context) ([]*user, error) {
	var users []*user
	for _, u := range up.mem {
		// we copy to avoid referring the same pointer that would get updated
		u := u
		users = append(users, &u)
	}

	return users, nil
}
