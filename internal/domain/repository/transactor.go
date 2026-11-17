package repository

import "context"

// Transactor runs fn inside a database transaction. Repositories called with
// the ctx passed to fn participate in that transaction; returning an error
// rolls everything back.
type Transactor interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
