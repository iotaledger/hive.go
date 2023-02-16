module github.com/iotaledger/hive.go/app

go 1.19

// Temporary replace until we are done with refactoring hive.go
replace github.com/iotaledger/hive.go/core => ../core

replace github.com/iotaledger/hive.go/runtime => ../runtime

require (
	github.com/hashicorp/go-version v1.6.0
	github.com/iotaledger/hive.go/core v1.0.0-rc.3
	github.com/iotaledger/hive.go/runtime v1.0.0-00010101000000-000000000000
	github.com/knadh/koanf v1.4.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cast v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.1
	github.com/tcnksm/go-latest v0.0.0-20170313132115-e3007ae9052e
	go.uber.org/dig v1.16.1
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/cockroachdb/errors v1.9.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ethereum/go-ethereum v1.10.26 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/getsentry/sentry-go v0.15.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-github v17.0.0+incompatible // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-rc.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/petermattis/goid v0.0.0-20221018141743-354ef7f2fd21 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.23.0 // indirect
	golang.org/x/crypto v0.2.0 // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
