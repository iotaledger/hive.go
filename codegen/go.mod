module github.com/iotaledger/hive.go/codegen

go 1.19

replace github.com/iotaledger/hive.go/core v0.0.0-unpublished => ../core

replace github.com/iotaledger/hive.go/ds v0.0.0-unpublished => ../ds

replace github.com/iotaledger/hive.go/lo v0.0.0-unpublished => ../lo

replace github.com/iotaledger/hive.go/serializer/v2 v2.0.0-unpublished => ../serializer

require (
	github.com/iotaledger/hive.go/lo v0.0.0-unpublished
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
)

require github.com/iotaledger/hive.go/constraints v0.0.0-20230216124949-dcd0bf545fea // indirect
