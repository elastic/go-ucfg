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

func MustFailUnpack(t testingT, cfg config, test interface{}) {
	if err := cfg.UnpackWithoutOptions(test); err == nil {
		t.Fatalf("expected failure, config:%s test:%s", spew.Sdump(cfg), spew.Sdump(test))
	}
}

func MustUnpack(t testingT, cfg config, test interface{}) {
	if err := cfg.UnpackWithoutOptions(test); err != nil {
		t.Fatalf("config:%s test:%s error:%v", spew.Sdump(cfg), spew.Sdump(test), err)
	}
}
