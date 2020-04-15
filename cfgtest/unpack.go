package cfgtest

import (
	"github.com/davecgh/go-spew/spew"
)

type testingT interface {
	Fatalf(format string, args ...interface{})
}

type config interface {
	UnpackWithoutOptions(to interface{}) error
}

// MustFailUnpack method fails the testing if unpacking passed.
func MustFailUnpack(t testingT, cfg config, test interface{}) {
	if err := cfg.UnpackWithoutOptions(test); err == nil {
		t.Fatalf("expected failure, config:%s test:%s", spew.Sdump(cfg), spew.Sdump(test))
	}
}

// MustUnpack method fails the testing if unpacking failed too.
func MustUnpack(t testingT, cfg config, test interface{}) {
	if err := cfg.UnpackWithoutOptions(test); err != nil {
		t.Fatalf("config:%s test:%s error:%v", spew.Sdump(cfg), spew.Sdump(test), err)
	}
}
