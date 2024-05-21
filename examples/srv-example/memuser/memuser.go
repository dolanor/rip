package memuser

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/dolanor/rip"
)

// start User Provider OMIT
type UserProvider struct {
	mem    map[int]*User
	logger *log.Logger
}

// end User Provider OMIT

func NewUserProvider(logger *log.Logger) *UserProvider {
	u := User{ID: 1, Name: "Jean", EmailAddress: "jean@example.com", BirthDate: time.Date(1900, time.November, 15, 0, 0, 0, 0, time.UTC)}
	return &UserProvider{
		mem: map[int]*User{
			u.ID: &u,
		},
		logger: logger,
	}
}

func (up *UserProvider) Create(ctx context.Context, u User) (User, error) {
	up.logger.Printf("SaveUser: %+v", u)
	id := rand.Intn(1000)
	u.ID = id

	up.mem[u.ID] = &u
	return u, nil
}

func (up UserProvider) Get(ctx context.Context, idString string) (User, error) {
	up.logger.Printf("GetUser: %+v", idString)

	if idString == "" {
		return User{}, nil
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		return User{}, err
	}

	u, ok := up.mem[id]
	if !ok {
		return User{}, rip.ErrNotFound
	}
	return *u, nil
}

func (up *UserProvider) Delete(ctx context.Context, idString string) error {
	up.logger.Printf("DeleteUser: %+v", idString)
	id, err := strconv.Atoi(idString)
	if err != nil {
		return err
	}

	_, ok := up.mem[id]
	if !ok {
		return rip.ErrNotFound
	}

	delete(up.mem, id)
	return nil
}

// start User Provider Update OMIT
func (up *UserProvider) Update(ctx context.Context, u User) error {
	up.logger.Printf("UpdateUser: %+v", u.ID)
	_, ok := up.mem[u.ID]
	if !ok {
		return rip.ErrNotFound
	}
	up.mem[u.ID] = &u

	return nil
}

// end User Provider Update OMIT

func (up UserProvider) List(ctx context.Context, offset, limit int) ([]User, error) {
	up.logger.Printf("ListUser")

	max := len(up.mem)
	if offset > max {
		offset = max
	}

	if offset+limit > max {
		limit = max - offset
	}

	var users []User
	for _, u := range up.mem {
		u := u
		users = append(users, *u)
	}

	return users[offset : offset+limit], nil
}
