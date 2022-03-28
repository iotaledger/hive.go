module github.com/iotaledger/hive.go

go 1.18

require (
	github.com/cockroachdb/errors v1.8.1
	github.com/cockroachdb/pebble v0.0.0-20220224165957-0e0d279abe38
	github.com/dgraph-io/badger/v2 v2.2007.4
	github.com/emirpasic/gods v1.12.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/websocket v1.5.0
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-20220309063734-061146d8ff30
	github.com/knadh/koanf v1.4.0
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
	github.com/ReneKroon/ttlcache/v2 v2.11.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.3.0 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
)

replace github.com/iotaledger/hive.go/serializer/v2 => ./serializer
