package domain

import "context"

type TransactionManager interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}
