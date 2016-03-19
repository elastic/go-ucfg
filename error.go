package ucfg

import (
	"errors"
	"fmt"
)

var (
	ErrMissing = errors.New("field name missing")

	ErrTypeNoArray = errors.New("field is no array")

	ErrTypeMismatch = errors.New("type mismatch")

	ErrKeyTypeNotString = errors.New("key must be a string")

	ErrIndexOutOfRange = errors.New("index out of range")

	ErrPointerRequired = errors.New("requires pointer for unpacking")

	ErrArraySizeMistach = errors.New("Array size mismatch")

	ErrExpectedObject = errors.New("expected object")

	ErrNilConfig = errors.New("config is nil")

	ErrNilValue = errors.New("unexpected nil value")

	ErrTODO = errors.New("TODO - implement me")
)

func raise(err error) error {
	// fmt.Println(string(debug.Stack()))
	return err
}

func errDuplicateKey(name string) error {
	return fmt.Errorf("duplicate field key '%v'", name)
}
