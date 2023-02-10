//go:build ignore

package main

import (
	"os"
	"strconv"

	"github.com/iotaledger/hive.go/core/codegen"
)

// This file is used to generate the variadic generic event implementations.
func main() {
	if len(os.Args) != 2 {
		panic("expected at least one argument (the amount of variadics to generate)")
	}

	paramCount, paramCountErr := strconv.Atoi(os.Args[1])
	panicOnError(paramCountErr)

	template := codegen.NewVariadic()
	panicOnError(template.Parse(os.Getenv("GOFILE")))
	panicOnError(template.Generate("variadic_generated.go", paramCount))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
