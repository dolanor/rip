package gormprovider

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dolanor/rip/internal/ripreflect"
)

func New[Ent any](db *gorm.DB, logger *slog.Logger) *gormEntityProvider[Ent] {
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

	res := ep.db.Create(&e)
	if res.Error != nil {
		return e, res.Error
	}

	return e, nil
}

func (ep *gormEntityProvider[Ent]) Delete(ctx context.Context, id string) error {
	ep.logger.Info("delete", "id", id)

	var e Ent
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
	ep.logger.Info("update", "id", id)
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

	ep.logger.Info("get", "id", id)

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
	ep.logger.Info("list")

	var ee []Ent

	tx := ep.db.
		Offset(offset).
		Limit(limit).
		Find(&ee)
	if tx.Error != nil {
		return ee, tx.Error
	}

	return ee, nil
}
