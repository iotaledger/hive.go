//go:build ignore

package main

import (
	"os"
	"strconv"

	"github.com/iotaledger/hive.go/core/codegen"
	"github.com/iotaledger/hive.go/core/generics/lo"
)

// This file is used to generate the variadic generic event implementations.
func main() {
	if len(os.Args) != 2 {
		panic("expected at least one argument (the amount of variadics to generate)")
	}

	template := codegen.NewVariadic()
	noError(template.Parse(os.Getenv("GOFILE")))
	noError(template.Generate("variadic_generated.go", lo.PanicOnErr(strconv.Atoi(os.Args[1]))))
}

func noError(err error) {
	if err != nil {
		panic(err)
	}
}
