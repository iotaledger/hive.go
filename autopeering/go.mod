module github.com/iotaledger/hive.go/autopeering

go 1.19

replace github.com/iotaledger/hive.go/core v0.0.0-unpublished => ../core

replace github.com/iotaledger/hive.go/kvstore v0.0.0-unpublished => ../kvstore

require (
	github.com/golang/protobuf v1.5.2
	github.com/iotaledger/hive.go/core v0.0.0-unpublished
	github.com/iotaledger/hive.go/kvstore v0.0.0-unpublished
	github.com/iotaledger/hive.go/lo v0.0.0-20230216132042-9c5c69b6d86c
	github.com/iotaledger/hive.go/runtime v0.0.0-20230216133807-1dcef29e3bc2
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.1
	go.uber.org/atomic v1.10.0
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/cockroachdb/errors v1.9.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ethereum/go-ethereum v1.10.26 // indirect
	github.com/getsentry/sentry-go v0.12.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/iotaledger/hive.go/constraints v0.0.0-20230216124949-dcd0bf545fea // indirect
	github.com/iotaledger/hive.go/ds v0.0.0-20230216133508-4294d334c92a // indirect
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-rc.1.0.20230216132042-9c5c69b6d86c // indirect
	github.com/iotaledger/hive.go/stringify v0.0.0-20230216132042-9c5c69b6d86c // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/oasisprotocol/ed25519 v0.0.0-20210505154701-76d8c688d86e // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/crypto v0.4.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
