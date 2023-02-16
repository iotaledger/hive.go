module github.com/iotaledger/hive.go/ds

go 1.19

replace github.com/iotaledger/hive.go/lo v0.0.0-unpublished => ../lo

replace github.com/iotaledger/hive.go/constraints v0.0.0-unpublished => ../constraints

replace github.com/iotaledger/hive.go/serializer/v2 v2.0.0-unpublished => ../serializer

require (
	github.com/emirpasic/gods v1.18.1
	github.com/iotaledger/hive.go/constraints v0.0.0-unpublished
	github.com/iotaledger/hive.go/lo v0.0.0-unpublished
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-unpublished
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.1
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ethereum/go-ethereum v1.10.26 // indirect
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
