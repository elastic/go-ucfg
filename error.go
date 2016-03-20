package ucfg

import (
	"errors"
	"fmt"
)

type Error interface {
	error
	Reason() error
}

type ValueError struct {
	reason error
	value  value
}

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

func (v ValueError) Error() string {
	return v.reason.Error()
}

func (v ValueError) Reason() error {
	return v.reason
}

/*
func raise(err error, value) error {
	// fmt.Println(string(debug.Stack()))
	return Error{err, value}
}
*/

func errDuplicateKey(name string) error {
	return fmt.Errorf("duplicate field key '%v'", name)
}

func raiseMissing(c *Config, field string) error {
	return ErrMissing
}

func raiseMissingArr(c *Config, field string, idx int) error {
	return ErrMissing
}

func raiseIndexOutOfBounds(c *Config, field string, idx int) error {
	return ErrIndexOutOfRange
}

func raiseInvalidTopLevelType(v interface{}) error {
	return ErrTypeMismatch
}

func raiseExpectedObject(cfg *Config, field string, v value) error {
	return ErrExpectedObject
}

func raiseKeyInvalidType() error {
	// most likely developers fault
	return ErrKeyTypeNotString
}

func raiseSquashNeedsObject() error {
	// most likely developers fault
	return ErrTypeMismatch
}

func raiseInlineNeedsObject() error {
	// most likely developers fault
	return ErrTypeMismatch
}

func raiseUnsupportedInputType() error {
	return ErrTypeMismatch
}

func raiseNil(reason error) error {
	// programmers error (passed unexpected nil pointer)
	return ErrNilValue
}

func raisePointerRequired() error {
	// developer did not pass pointer, unpack target is not settable
	return ErrPointerRequired
}

func raiseToTypeNotSupported() error {
	return ErrTODO
}

func raiseArraySize() error {
	return ErrArraySizeMistach
}

func raiseValueNotBool() error {
	return ErrTypeMismatch
}

func raiseValueNotString() error {
	return ErrTypeMismatch
}

func raiseValueNotInt() error {
	return ErrTypeMismatch
}

func raiseValueNotFloat() error {
	return ErrTypeMismatch
}

func raiseValueNotObject() error {
	return ErrTypeMismatch
}
