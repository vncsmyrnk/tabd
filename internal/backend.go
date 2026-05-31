package internal

import "context"

type Options struct {
	Name       string
	FilePath   string
	MountPoint string
	StowPath   string
}

func (o Options) Valid() bool {
	return o.Name != "" && o.FilePath != ""
}

type Backend interface {
	UnlockVault(ctx context.Context, vault, password string) error
	LockVault(ctx context.Context, vault string) error
	GenerateService(ctx context.Context, options Options) error
}
