package godbprovider

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/samonzeweb/godb"

	"github.com/dolanor/rip/internal/ripreflect"
)

func New[Ent any](db *godb.DB, logger *slog.Logger) *godbEntityProvider[Ent] {
	if logger == nil {
		logger = slog.Default()
	}
	var e Ent
	idFieldName := ripreflect.FieldIDName(e)
	if idFieldName == ripreflect.MissingIDField {
		// we should stop here as the entity is not valid
		panic("no ID field")
	}

	return &godbEntityProvider[Ent]{
		db:     db,
		logger: logger,
	}
}

type godbEntityProvider[Ent any] struct {
	db *godb.DB

	logger *slog.Logger
}

func (ep *godbEntityProvider[Ent]) Create(ctx context.Context, e Ent) (Ent, error) {
	id, err := ripreflect.GetID(e)
	ep.logger.Info("create", "id", id)
	if err != nil {
		return e, err
	}

	if id == "" {
		uuid, err := uuid.NewV7()
		if err != nil {
			return e, errors.New("can not generate unique id")
		}

		err = ripreflect.SetID(&e, uuid.String())
		if err != nil {
			return e, errors.New("can not set unique id")
		}

	}

	err = ep.db.Insert(&e).Do()
	if err != nil {
		return e, err
	}

	return e, nil
}

func (ep *godbEntityProvider[Ent]) Delete(ctx context.Context, id string) error {
	ep.logger.Info("delete", "id", id)

	var e Ent
	err := ripreflect.SetID(&e, id)
	if err != nil {
		return err
	}

	_, err = ep.db.Delete(&e).Do()
	if err != nil {
		return err
	}

	return nil
}

func (ep *godbEntityProvider[Ent]) Update(ctx context.Context, e Ent) error {
	id, err := ripreflect.GetID(e)
	ep.logger.Info("update", "id", id)
	if err != nil {
		return err
	}

	err = ep.db.Update(&e).Do()
	if err != nil {
		return err
	}
	return nil
}

func (ep *godbEntityProvider[Ent]) Get(ctx context.Context, id string) (Ent, error) {
	var e Ent
	if id == "" {
		return e, nil
	}

	ep.logger.Info("get", "id", id)

	idFieldName := ripreflect.FieldIDName(e)
	if idFieldName == ripreflect.MissingIDField {
		return e, errors.New("no ID field")
	}

	err := ep.db.Select(&e).
		Where(fmt.Sprintf("%s = ?", idFieldName), id).
		Do()
	if err != nil {
		return e, err
	}

	return e, nil
}

func (ep *godbEntityProvider[Ent]) List(ctx context.Context, offset, limit int) ([]Ent, error) {
	ep.logger.Info("list")

	var ee []Ent

	err := ep.db.Select(&ee).Do()
	if err != nil {
		return ee, err
	}

	return ee, nil
}
