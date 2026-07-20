//go:build darwin

package keyctl

import (
	"errors"
	"time"
)

const (
	DefaultKeySafe = "keysafe"
)

var (
	ErrNoKey          = errors.New("Required key not available")
	ErrKeyExpired     = errors.New("Key has expired")
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

func (k *KeySafe) SetTimeout(name string, timeout time.Duration) error {
	return ErrNotImplemented
}

func (k *KeySafe) List() ([]string, error) {
	return nil, ErrNotImplemented
}

func (k *KeySafe) Reap() (int, error) {
	return 0, ErrNotImplemented
}

func (k *KeySafe) Clear() error {
	return ErrNotImplemented
}

func (k *KeySafe) Del(name string) error {
	return ErrNotImplemented
}
