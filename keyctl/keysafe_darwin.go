//go:build darwin

package keyctl

import (
	"errors"
)

const (
	DefaultKeySafe = "keysafe"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type KeySafe struct{}

func NewKeySafe(kname string) (*KeySafe, error) {
	return nil, ErrNotImplemented
}

func (k *KeySafe) Get(name string) ([]byte, error) {
	return nil, ErrNotImplemented
}
func (k *KeySafe) Set(name string, value []byte) error {
	return ErrNotImplemented
}

func (k *KeySafe) List() ([]string, error) {
	return nil, ErrNotImplemented
}

func (k *KeySafe) Clear() error {
	return ErrNotImplemented
}

func (k *KeySafe) Del(name string) error {
	return ErrNotImplemented
}
