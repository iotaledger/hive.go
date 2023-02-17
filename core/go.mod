module github.com/iotaledger/hive.go/core

go 1.19

replace github.com/iotaledger/hive.go/runtime => ../runtime

require (
	github.com/cockroachdb/errors v1.9.1
	github.com/cockroachdb/pebble v0.0.0-20230209160836-829675f94811
	github.com/dgraph-io/badger/v2 v2.2007.4
	github.com/emirpasic/gods v1.18.1
	github.com/ethereum/go-ethereum v1.11.1
	github.com/golang/protobuf v1.5.2
	github.com/iancoleman/orderedmap v0.2.0
	github.com/iotaledger/grocksdb v1.7.5-0.20221128103803-fcdb79760195
	github.com/iotaledger/hive.go/runtime v0.0.0-00010101000000-000000000000
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-rc.1
	github.com/jellydator/ttlcache/v2 v2.11.1
	github.com/kr/text v0.2.0
	github.com/libp2p/go-libp2p v0.23.4
	github.com/mr-tron/base58 v1.2.0
	github.com/oasisprotocol/ed25519 v0.0.0-20210505154701-76d8c688d86e
	github.com/pelletier/go-toml/v2 v2.0.5
	github.com/pkg/errors v0.9.1
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/stretchr/testify v1.8.1
	go.dedis.ch/kyber/v3 v3.0.14
	go.uber.org/atomic v1.10.0
	go.uber.org/zap v1.23.0
	golang.org/x/crypto v0.6.0
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
	google.golang.org/protobuf v1.28.1
	nhooyr.io/websocket v1.8.7
)

require (
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/dgraph-io/ristretto v0.0.3-0.20200630154024-f66de99634de // indirect
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/getsentry/sentry-go v0.18.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.15.15 // indirect
	github.com/klauspost/cpuid/v2 v2.1.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/libp2p/go-openssl v0.1.0 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/petermattis/goid v0.0.0-20221018141743-354ef7f2fd21 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	go.dedis.ch/fixbuf v1.0.3 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/exp v0.0.0-20230206171751-46f607a40771 // indirect
	golang.org/x/net v0.6.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
