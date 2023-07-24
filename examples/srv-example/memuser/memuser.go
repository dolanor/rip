package memuser

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/dolanor/rip"
)

type User struct {
	ID        int       `json:"id" xml:"id"`
	Name      string    `json:"name" xml:"name"`
	BirthDate time.Time `json:"birth_date" xml:"birth_date"`
}

func (u User) IDString() string {
	return strconv.Itoa(u.ID)
}

func (u *User) IDFromString(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	u.ID = n
	return nil
}

type UserProvider struct {
	mem map[int]*User
}

func NewUserProvider() *UserProvider {
	u := User{ID: 1, Name: "Jean"}
	return &UserProvider{
		mem: map[int]*User{
			u.ID: &u,
		},
	}
}

func (up *UserProvider) Create(ctx context.Context, u *User) (*User, error) {
	log.Printf("SaveUser: %+v", *u)
	id := rand.Intn(1000)
	u.ID = id

	up.mem[u.ID] = u
	return u, nil
}

func (up UserProvider) Get(ctx context.Context, res rip.IdentifiableResource) (*User, error) {
	log.Printf("GetUser: %+v", res.IDString())
	if res.IDString() == "new" {
		return &User{}, nil
	}
	id, err := strconv.Atoi(res.IDString())
	if err != nil {
		return nil, err
	}

	u, ok := up.mem[id]
	if !ok {
		return &User{}, rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}
	return u, nil
}

func (up *UserProvider) Delete(ctx context.Context, res rip.IdentifiableResource) error {
	log.Printf("DeleteUser: %+v", res.IDString())
	id, err := strconv.Atoi(res.IDString())
	if err != nil {
		return err
	}

	_, ok := up.mem[id]
	if !ok {
		return rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}

	delete(up.mem, id)
	return nil
}

func (up *UserProvider) Update(ctx context.Context, u *User) error {
	log.Printf("UpdateUser: %+v", u.IDString())
	_, ok := up.mem[u.ID]
	if !ok {
		return rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}
	up.mem[u.ID] = u

	return nil
}

func (up UserProvider) ListAll(ctx context.Context) ([]*User, error) {
	log.Printf("ListAllUser")
	var users []*User
	for _, u := range up.mem {
		u := u
		users = append(users, u)
	}
	return users, nil
}
