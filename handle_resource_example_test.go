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

	http.HandleFunc(HandleEntities("/users/", up, WithCodecs(json.Codec, html.NewEntityCodec("/users/"))))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

type user struct {
	Name         string    `rip:"id" json:"name" xml:"name"`
	EmailAddress string    `json:"email_address" xml:"email_address"`
	BirthDate    time.Time `json:"birth_date" xml:"birth_date"`
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

func (up UserProvider) Get(ctx context.Context, id string) (*user, error) {
	u, ok := up.mem[id]
	if !ok {
		return &user{}, ErrNotFound
	}
	return &u, nil
}

func (up *UserProvider) Delete(ctx context.Context, id string) error {
	_, ok := up.mem[id]
	if !ok {
		return ErrNotFound
	}

	delete(up.mem, id)
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

func (up *UserProvider) List(ctx context.Context, offset, limit int) ([]*user, error) {
	var users []*user

	max := len(up.mem)
	if offset > max {
		offset = max
	}

	if offset+limit > max {
		limit = max - offset
	}

	for _, u := range up.mem {
		// we copy to avoid referring the same pointer that would get updated
		u := u
		users = append(users, &u)
	}

	return users[offset : offset+limit], nil
}
