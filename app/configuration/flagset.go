package configuration

import (
	flag "github.com/spf13/pflag"
)

// NewUnsortedFlagSet creates a new unsorted FlagSet.
func NewUnsortedFlagSet(name string, errorHandling flag.ErrorHandling) *flag.FlagSet {
	flagset := flag.NewFlagSet(name, errorHandling)
	flagset.SortFlags = false

	return flagset
}

// HasFlag checks if a flag with the given name exists in the flagset.
func HasFlag(flagSet *flag.FlagSet, name string) bool {
	has := false
	flagSet.Visit(func(f *flag.Flag) {
		if f.Name == name {
			has = true
		}
	})

	return has
}

// ParseFlagSets adds the given flag sets to flag.CommandLine and then parses them.
func ParseFlagSets(flagSets []*flag.FlagSet) {
	for _, flagSet := range flagSets {
		flag.CommandLine.AddFlagSet(flagSet)
	}
	flag.Parse()
}

// HideFlags hides all non essential flags from the help/usage text.
func HideFlags(flagSets []*flag.FlagSet, nonHiddenFlag []string) {
	nonHiddenFlagsMap := make(map[string]struct{})
	for _, flag := range nonHiddenFlag {
		nonHiddenFlagsMap[flag] = struct{}{}
	}

	flag.VisitAll(func(f *flag.Flag) {
		_, notHidden := nonHiddenFlagsMap[f.Name]
		f.Hidden = !notHidden
	})

	for _, flagset := range flagSets {
		flagset.VisitAll(func(f *flag.Flag) {
			_, notHidden := nonHiddenFlagsMap[f.Name]
			f.Hidden = !notHidden
		})
	}
}
