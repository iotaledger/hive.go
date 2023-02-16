module github.com/iotaledger/hive.go/kvstore

go 1.19

replace github.com/iotaledger/hive.go/core v0.0.0-unpublished => ../core

replace github.com/iotaledger/hive.go/ds v0.0.0-unpublished => ../ds

replace github.com/iotaledger/hive.go/runtime v0.0.0-unpublished => ../runtime

replace github.com/iotaledger/hive.go/constraints v0.0.0-unpublished => ../constraints

replace github.com/iotaledger/hive.go/stringify v0.0.0-unpublished => ../stringify

replace github.com/iotaledger/hive.go/serializer/v2 v2.0.0-unpublished => ../serializer

require (
	github.com/cockroachdb/pebble v0.0.0-20221111210721-1bda21f14fc2
	github.com/dgraph-io/badger/v2 v2.2007.4
	github.com/iotaledger/grocksdb v1.7.5-0.20221128103803-fcdb79760195
	github.com/iotaledger/hive.go/constraints v0.0.0-unpublished
	github.com/iotaledger/hive.go/core v0.0.0-unpublished
	github.com/iotaledger/hive.go/ds v0.0.0-unpublished
	github.com/iotaledger/hive.go/runtime v0.0.0-unpublished
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-unpublished
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.1
	go.uber.org/atomic v1.10.0
)

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cockroachdb/errors v1.9.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/ristretto v0.0.3-0.20200630154024-f66de99634de // indirect
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/getsentry/sentry-go v0.15.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/iotaledger/hive.go/stringify v0.0.0-unpublished // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/petermattis/goid v0.0.0-20221018141743-354ef7f2fd21 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.2.1-0.20210607210712-147c58e9608a // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	golang.org/x/exp v0.0.0-20220916125017-b168a2c6b86b // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
