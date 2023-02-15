package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/hive.go/codegen/variadic"
)

// main is the entry point of the variadic code generator.
func main() {
	if len(os.Args) < 4 {
		printUsage("not enough parameters")
	}

	minParamsCount, err := strconv.Atoi(os.Args[1])
	if err != nil {
		printUsage("minParamsCount (1st parameter) must be an integer")
	}

	maxParamsCount, err := strconv.Atoi(os.Args[2])
	if err != nil {
		printUsage("maxParamsCount (2nd parameter) must be an integer")
	}

	template := variadic.New()
	panicOnErr(template.Parse(os.Getenv("GOFILE")))
	panicOnErr(template.Generate(os.Args[3], minParamsCount, maxParamsCount))
}

// printUsage prints the usage of the variadic code generator in case of an error.
func printUsage(errorMsg string) {
	_, _ = fmt.Fprintf(os.Stderr, "Error:\t%s\n\n", errorMsg)
	_, _ = fmt.Fprintf(os.Stderr, "Usage of variadic:\n")
	_, _ = fmt.Fprintf(os.Stderr, "\tvariadic [minParamsCount] [maxParamsCount] [outputFile]\n")

	os.Exit(2)
}

// panicOnErr panics if the given error is not nil.
func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
