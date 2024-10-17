package mapprovider

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strconv"
	"sync"

	"github.com/google/uuid"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/internal/ripreflect"
)

func New[Ent any](logger *slog.Logger) entityMapProvider[Ent] {
	return entityMapProvider[Ent]{
		store:  map[string]Ent{},
		logger: logger,
	}
}

type entityMapProvider[Ent any] struct {
	mu             sync.Mutex
	store          map[string]Ent
	listCacheFresh bool

	listCache []Ent

	logger *slog.Logger
}

func (dp *entityMapProvider[Ent]) Create(ctx context.Context, d Ent) (Ent, error) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	id, err := ripreflect.GetID(d)
	dp.logger.Info("create", "id", id)
	if err != nil {
		return d, err
	}

	if id == "" {
		uuid, err := uuid.NewV7()
		if err != nil {
			return d, errors.New("can not generate unique id")
		}

		err = ripreflect.SetID(&d, uuid.String())
		if err != nil {
			return d, errors.New("can not set unique id")
		}

	}

	dp.store[id] = d
	dp.listCacheFresh = false
	return d, nil
}

func (dp *entityMapProvider[Ent]) Delete(ctx context.Context, id string) error {
	dp.logger.Info("delete", "id", id)
	dp.mu.Lock()
	defer dp.mu.Unlock()

	delete(dp.store, id)
	dp.listCacheFresh = false

	return nil
}

func (dp *entityMapProvider[Ent]) Update(ctx context.Context, e Ent) error {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	id, err := ripreflect.GetID(e)
	dp.logger.Info("update", "id", id)
	if err != nil {
		return err
	}

	dp.store[id] = e
	dp.listCacheFresh = false
	return nil
}

func (dp *entityMapProvider[Ent]) Get(ctx context.Context, id string) (Ent, error) {
	var e Ent
	if id == "" {
		return e, nil
	}

	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.logger.Info("get", "id", id)
	e, ok := dp.store[id]
	if !ok {
		return e, rip.ErrNotFound
	}
	return e, nil
}

func (dp *entityMapProvider[Ent]) List(ctx context.Context, offset, limit int) ([]Ent, error) {
	dp.logger.Info("list")
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if dp.listCacheFresh {
		return dp.listCache, nil
	}

	var dd []Ent
	for _, v := range dp.store {
		dd = append(dd, v)
	}

	slices.SortFunc(dd, func(a, b Ent) int {
		idA, err := ripreflect.GetID(a)
		if err != nil {
			return -1
		}

		idB, err := ripreflect.GetID(b)
		if err != nil {
			return -1
		}

		comparison, ok := compareAsNumbers(idA, idB)
		if ok {
			return comparison
		}

		if idA < idB {
			return -1
		} else if idA > idB {
			return 1
		}

		return 0
	})
	dp.listCacheFresh = true
	dp.listCache = dd

	return dp.listCache, nil
}

func compareAsNumbers(idA, idB string) (comparison int, ok bool) {
	var isNumber bool
	idAInt, err := strconv.Atoi(idA)
	if err == nil {
		isNumber = true
	}

	idBInt, err := strconv.Atoi(idB)
	if err == nil && isNumber {
		isNumber = true
	}

	if isNumber {
		return idAInt - idBInt, true
	}
	return 0, false
}
