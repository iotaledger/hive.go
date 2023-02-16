module github.com/iotaledger/hive.go/runtime

go 1.19

replace github.com/iotaledger/hive.go/core v0.0.0-unpublished => ../core

replace github.com/iotaledger/hive.go/ds v0.0.0-unpublished => ../ds

replace github.com/iotaledger/hive.go/constraints v0.0.0-unpublished => ../constraints

replace github.com/iotaledger/hive.go/lo v0.0.0-unpublished => ../lo

replace github.com/iotaledger/hive.go/stringify v0.0.0-unpublished => ../stringify

replace github.com/iotaledger/hive.go/serializer/v2 v2.0.0-unpublished => ../serializer

require (
	github.com/iotaledger/hive.go/constraints v0.0.0-unpublished
	github.com/iotaledger/hive.go/core v0.0.0-unpublished
	github.com/iotaledger/hive.go/ds v0.0.0-unpublished
	github.com/iotaledger/hive.go/lo v0.0.0-unpublished
	github.com/iotaledger/hive.go/stringify v0.0.0-unpublished
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/stretchr/testify v1.8.1
	go.uber.org/atomic v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
