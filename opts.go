package ucfg

type Option func(*options)

type options struct {
	tag          string
	validatorTag string
	pathSep      string
	meta         *Meta
	varexp       bool
}

func StructTag(tag string) Option {
	return func(o *options) {
		o.tag = tag
	}
}

func ValidatorTag(tag string) Option {
	return func(o *options) {
		o.validatorTag = tag
	}
}

func PathSep(sep string) Option {
	return func(o *options) {
		o.pathSep = sep
	}
}

func MetaData(meta Meta) Option {
	return func(o *options) {
		o.meta = &meta
	}
}

var VarExp Option = func(o *options) {
	o.varexp = true
}

func makeOptions(opts []Option) options {
	o := options{
		tag:          "config",
		validatorTag: "validate",
		pathSep:      "", // no separator by default
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
