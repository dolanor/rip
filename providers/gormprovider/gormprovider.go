package gormprovider

import (
	"context"
	"errors"
	"log/slog"
	"reflect"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dolanor/rip/internal/ripreflect"
)

func New[Ent any](db *gorm.DB, logger *slog.Logger) *gormEntityProvider[Ent] {
	if logger == nil {
		logger = slog.Default()
	}
	var e Ent
	idFieldName := ripreflect.FieldIDName(e)
	if idFieldName == ripreflect.MissingIDField {
		// we should stop here as the entity is not valid
		panic("no ID field")
	}

	return &gormEntityProvider[Ent]{
		db:     db,
		logger: logger,
	}
}

type gormEntityProvider[Ent any] struct {
	db *gorm.DB

	logger *slog.Logger
}

func (ep *gormEntityProvider[Ent]) Create(ctx context.Context, e Ent) (Ent, error) {
	id, err := ripreflect.GetID(e)
	defer func() { ep.logger.Info("create", "entity", reflect.TypeOf(e).Name(), "id", id) }()
	if err != nil {
		return e, err
	}

	if id == "" {
		uuid, err := uuid.NewV7()
		if err != nil {
			return e, errors.New("can not generate unique id")
		}
		id = uuid.String()

		err = ripreflect.SetID(&e, uuid.String())
		if err != nil {
			return e, errors.New("can not set unique id")
		}

	}

	res := ep.db.Create(&e)
	if res.Error != nil {
		return e, res.Error
	}

	return e, nil
}

func (ep *gormEntityProvider[Ent]) Delete(ctx context.Context, id string) error {
	var e Ent
	ep.logger.Info("delete", "entity", reflect.TypeOf(e).Name(), "id", id)
	err := ripreflect.SetID(&e, id)
	if err != nil {
		return err
	}

	tx := ep.db.Delete(&e)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (ep *gormEntityProvider[Ent]) Update(ctx context.Context, e Ent) error {
	id, err := ripreflect.GetID(e)
	ep.logger.Info("update", "entity", reflect.TypeOf(e).Name(), "id", id)
	if err != nil {
		return err
	}

	tx := ep.db.Save(&e)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (ep *gormEntityProvider[Ent]) Get(ctx context.Context, id string) (Ent, error) {
	var e Ent
	if id == "" {
		return e, nil
	}

	ep.logger.Info("get", "entity", reflect.TypeOf(e).Name(), "id", id)

	idFieldName := ripreflect.FieldIDName(e)
	if idFieldName == ripreflect.MissingIDField {
		return e, errors.New("no ID field")
	}

	tx := ep.db.First(&e)
	if tx.Error != nil {
		return e, tx.Error
	}

	return e, nil
}

func (ep *gormEntityProvider[Ent]) List(ctx context.Context, offset, limit int) ([]Ent, error) {
	var e Ent
	var ee []Ent
	defer func() {
		ep.logger.Info("list", "entity", reflect.TypeOf(e).Name(), "offset", offset, "limit", limit, "size", len(ee))
	}()

	tx := ep.db.
		Offset(offset).
		Limit(limit).
		Find(&ee)
	if tx.Error != nil {
		return ee, tx.Error
	}

	return ee, nil
}
