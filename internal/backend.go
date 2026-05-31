package internal

import "context"

type Backend interface {
	UnlockVault(ctx context.Context, password string) error
	LockVault(ctx context.Context) error
}
