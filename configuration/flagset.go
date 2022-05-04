package configuration

import (
	flag "github.com/spf13/pflag"
)

func NewUnsortedFlagSet(name string, errorHandling flag.ErrorHandling) *flag.FlagSet {
	flagset := flag.NewFlagSet(name, errorHandling)
	flagset.SortFlags = false
	return flagset
}
