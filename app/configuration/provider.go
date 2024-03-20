package configuration

import (
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/pflag"

	"github.com/iotaledger/hive.go/ierrors"
)

// lowerPosflag implements a pflag command line provider.
type lowerPosflag struct {
	delim   string
	flagset *pflag.FlagSet
	ko      *koanf.Koanf
}

// lowerPosflagProvider returns a commandline flags provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
//
// It takes an optional (but recommended) Koanf instance to see if the
// flags defined have been set from other providers, for instance,
// a config file. If they are not, then the default values of the flags
// are merged. If they do exist, the flag values are not merged but only
// the values that have been explicitly set in the command line are merged.
func lowerPosflagProvider(f *pflag.FlagSet, delim string, ko *koanf.Koanf) *lowerPosflag {
	return &lowerPosflag{
		flagset: f,
		delim:   delim,
		ko:      ko,
	}
}

// Read reads the flag variables and returns a nested conf map.
func (p *lowerPosflag) Read() (map[string]interface{}, error) {
	mp := make(map[string]interface{})
	p.flagset.VisitAll(func(f *pflag.Flag) {
		// If no value was explicitly set in the command line,
		// check if the default value should be used.
		if !f.Changed {
			if p.ko != nil {
				if p.ko.Exists(strings.ToLower(f.Name)) {
					return
				}
			} else {
				return
			}
		}

		mp[strings.ToLower(f.Name)] = posflag.FlagVal(p.flagset, f)
	})

	return maps.Unflatten(mp, p.delim), nil
}

// ReadBytes is not supported by the env koanf.
func (p *lowerPosflag) ReadBytes() ([]byte, error) {
	return nil, ierrors.New("pflag provider does not support this method")
}

// Watch is not supported.
//
//nolint:revive
func (p *lowerPosflag) Watch(cb func(event interface{}, err error)) error {
	return ierrors.New("posflag provider does not support this method")
}
