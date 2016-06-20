package flag

import (
	goflag "flag"

	"github.com/urso/ucfg"
	"github.com/urso/ucfg/json"
	"github.com/urso/ucfg/yaml"
)

func ConfigVar(
	set *goflag.FlagSet,
	def *ucfg.Config,
	name string,
	usage string,
	opts ...ucfg.Option,
) *ucfg.Config {
	v := NewFlagKeyValue(def, true, opts...)
	registerFlag(set, v, name, usage)
	return v.Config()
}

func Config(
	set *goflag.FlagSet,
	name string,
	usage string,
	opts ...ucfg.Option,
) *ucfg.Config {
	return ConfigVar(set, nil, name, usage, opts...)
}

func ConfigFilesVar(
	set *goflag.FlagSet,
	def *ucfg.Config,
	name string,
	usage string,
	extensions map[string]FileLoader,
	opts ...ucfg.Option,
) *ucfg.Config {
	v := NewFlagFiles(def, extensions, opts...)
	registerFlag(set, v, name, usage)
	return v.Config()
}

func ConfigFiles(
	set *goflag.FlagSet,
	name string,
	usage string,
	extensions map[string]FileLoader,
	opts ...ucfg.Option,
) *ucfg.Config {
	return ConfigFilesVar(set, nil, name, usage, extensions, opts...)
}

func ConfigYAMLFilesVar(
	set *goflag.FlagSet,
	def *ucfg.Config,
	name string,
	usage string,
	opts ...ucfg.Option,
) *ucfg.Config {
	exts := map[string]FileLoader{"": yaml.NewConfigWithFile}
	return ConfigFilesVar(set, def, name, usage, exts, opts...)
}

func ConfigYAMLFiles(
	set *goflag.FlagSet,
	name string,
	usage string,
	opts ...ucfg.Option,
) *ucfg.Config {
	return ConfigYAMLFilesVar(set, nil, name, usage, opts...)
}

func ConfigJSONFilesVar(
	set *goflag.FlagSet,
	def *ucfg.Config,
	name string,
	usage string,
	opts ...ucfg.Option,
) *ucfg.Config {
	exts := map[string]FileLoader{"": json.NewConfigWithFile}
	return ConfigFilesVar(set, def, name, usage, exts, opts...)
}

func ConfigJSONFiles(
	set *goflag.FlagSet,
	name string,
	usage string,
	opts ...ucfg.Option,
) *ucfg.Config {
	return ConfigJSONFilesVar(set, nil, name, usage, opts...)
}

func ConfigFilesExtsVar(
	set *goflag.FlagSet,
	def *ucfg.Config,
	name string,
	usage string,
	opts ...ucfg.Option,
) *ucfg.Config {
	exts := map[string]FileLoader{
		".yaml": yaml.NewConfigWithFile,
		".yml":  yaml.NewConfigWithFile,
		".json": json.NewConfigWithFile,
	}
	return ConfigFilesVar(set, def, name, usage, exts, opts...)
}

func ConfigFilesExts(
	set *goflag.FlagSet,
	name string,
	usage string,
	opts ...ucfg.Option,
) *ucfg.Config {
	return ConfigFilesExtsVar(set, nil, name, usage, opts...)
}

func registerFlag(set *goflag.FlagSet, v goflag.Value, name, usage string) {
	if set != nil {
		set.Var(v, name, usage)
	} else {
		goflag.Var(v, name, usage)
	}
}
