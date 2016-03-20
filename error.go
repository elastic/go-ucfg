package ucfg

import (
	"errors"
	"fmt"
	"runtime/debug"
)

type Error interface {
	error
	Reason() error
	Class() error
	Trace() string // optional stack trace
}

type baseError struct {
	reason error
	class  error
}

type criticalError struct {
	baseError
	trace string
}

type pathError struct {
	baseError
	meta *Meta
	path string
}

type ValueError struct {
	reason error
	value  value
}

// error Reasons
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

// error classes
var (
	ErrConfig         = errors.New("Configuration error")
	ErrImplementation = errors.New("Implementation error")
	ErrUnknown        = errors.New("Unspecified")
)

func (e baseError) Error() string {
	return e.reason.Error()
}

func (e baseError) Reason() error {
	return e.reason
}

func (e baseError) Class() error {
	return e.class
}

func (e baseError) Trace() string {
	return ""
}

func (v ValueError) Error() string {
	return v.reason.Error()
}

func (v ValueError) Reason() error {
	return v.reason
}

func raiseErr(reason error) Error {
	return baseError{
		reason: reason,
		class:  ErrConfig,
	}
}

func raiseImplErr(reason error) Error {
	return baseError{
		reason: reason,
		class:  ErrImplementation,
	}
}

func raiseCritical(reason error) Error {
	return criticalError{
		baseError{reason, ErrImplementation},
		string(debug.Stack()),
	}
}

func raisePathErr(reason error, meta *Meta, path string) Error {
	return pathError{
		baseError{reason, ErrConfig},
		meta,
		path,
	}
}

func raiseDuplicateKey(cfg *Config, name string) Error {
	return raiseErr(
		fmt.Errorf("duplicate field key '%v'", name))
}

func raiseMissing(c *Config, field string) Error {
	// error reading field from config, as missing in c
	return raisePathErr(ErrMissing, c.metadata, c.PathOf(field, "."))
}

func raiseMissingArr(arr *cfgArray, idx int) Error {
	path := fmt.Sprintf("%v.%v", arr.ctx.path("."), idx)
	return raisePathErr(ErrMissing, arr.meta(), path)
}

func raiseIndexOutOfBounds(c *Config, field string, idx int, value value) Error {
	return raiseErr(ErrIndexOutOfRange)
}

func raiseInvalidTopLevelType(v interface{}) Error {
	// t := chaseTypePointers(chaseValue(reflect.ValueOf(v)).Type())
	// return ErrTypeMismatch
	return raiseCritical(ErrTypeMismatch)
}

func raiseKeyInvalidType() Error {
	// most likely developers fault
	return raiseCritical(ErrKeyTypeNotString)
}

func raiseSquashNeedsObject() Error {
	// most likely developers fault
	return raiseCritical(ErrTypeMismatch)
}

func raiseInlineNeedsObject() Error {
	// most likely developers fault
	return raiseCritical(ErrTypeMismatch)
}

func raiseUnsupportedInputType() Error {
	return raiseCritical(ErrTypeMismatch)
}

func raiseNil(reason error) Error {
	// programmers error (passed unexpected nil pointer)
	return raiseCritical(ErrNilValue)
}

func raisePointerRequired() Error {
	// developer did not pass pointer, unpack target is not settable
	return raiseCritical(ErrPointerRequired)
}

func raiseToTypeNotSupported() Error {
	return raiseCritical(ErrTODO)
}

func raiseArraySize() Error {
	return raiseErr(ErrArraySizeMistach)
}

func raiseExpectedObject(cfg *Config, field string, v value) Error {
	return raiseErr(ErrExpectedObject)
}

func raiseConversion(v value, err error, to string) Error {
	ctx := v.Context()
	path := ctx.path(".")
	return raisePathErr(err, v.meta(), path)
}
