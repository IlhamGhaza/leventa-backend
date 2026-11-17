package postgres

import (
	"context"

	"github.com/leventa-backend/internal/domain/repository"
	"gorm.io/gorm"
)

type txKey struct{}

type transactor struct {
	db *gorm.DB
}

// NewTransactor creates a Transactor backed by GORM.
func NewTransactor(db *gorm.DB) repository.Transactor {
	return &transactor{db: db}
}

// Do runs fn in a transaction. Nested calls reuse the outer transaction.
func (t *transactor) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// conn returns the transaction stored in ctx, or db when none is active.
func conn(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}
