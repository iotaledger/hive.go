module github.com/iotaledger/hive.go

go 1.18

require (
	github.com/cockroachdb/errors v1.8.1
	github.com/cockroachdb/pebble v0.0.0-20220224165957-0e0d279abe38
	github.com/dgraph-io/badger/v2 v2.2007.4
	github.com/emirpasic/gods v1.12.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/websocket v1.5.0
	github.com/knadh/koanf v1.4.0
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-20220309063734-061146d8ff30
	github.com/kr/text v0.2.0
	github.com/linxGnu/grocksdb v1.6.46
	github.com/mr-tron/base58 v1.2.0
	github.com/oasisprotocol/ed25519 v0.0.0-20210505154701-76d8c688d86e
	github.com/panjf2000/ants/v2 v2.4.8
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.dedis.ch/kyber/v3 v3.0.13
	go.etcd.io/bbolt v1.3.6
	go.uber.org/atomic v1.9.0
	go.uber.org/dig v1.13.0
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20190617123548-eb05cc24525f // indirect
	github.com/cockroachdb/redact v1.0.8 // indirect
	github.com/cockroachdb/sentry-go v0.6.1-cockroachdb.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/ristretto v0.0.3-0.20200630154024-f66de99634de // indirect
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/klauspost/compress v1.12.3 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.dedis.ch/fixbuf v1.0.3 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/exp v0.0.0-20200513190911-00229845015e // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210909193231-528a39cd75f3 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

require (
	github.com/ReneKroon/ttlcache/v2 v2.11.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.3.0 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
)

replace (
	"github.com/iotaledger/hive.go/serializer/v2" => ./serializer
)
