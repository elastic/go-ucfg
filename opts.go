package ucfg

type Option func(*options)

type options struct {
	tag     string
	pathSep string
}

func StructTag(tag string) Option {
	return func(o *options) {
		o.tag = tag
	}
}

func PathSep(sep string) Option {
	return func(o *options) {
		o.pathSep = sep
	}
}

func makeOptions(opts []Option) options {
	o := options{
		tag:     "config",
		pathSep: "", // no separator by default
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
