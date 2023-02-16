module github.com/iotaledger/hive.go/app

go 1.19

replace github.com/iotaledger/hive.go/core v0.0.0-unpublished => ../core

replace github.com/iotaledger/hive.go/runtime v0.0.0-unpublished => ../runtime

require (
	github.com/hashicorp/go-version v1.6.0
	github.com/iotaledger/hive.go/core v0.0.0-unpublished
	github.com/iotaledger/hive.go/runtime v0.0.0-unpublished
	github.com/knadh/koanf v1.4.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cast v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.1
	github.com/tcnksm/go-latest v0.0.0-20170313132115-e3007ae9052e
	go.uber.org/atomic v1.10.0
	go.uber.org/dig v1.16.1
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/google/go-github v17.0.0+incompatible // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/iotaledger/hive.go/constraints v0.0.0-20230216124949-dcd0bf545fea // indirect
	github.com/iotaledger/hive.go/ds v0.0.0-20230216133508-4294d334c92a // indirect
	github.com/iotaledger/hive.go/lo v0.0.0-20230216132042-9c5c69b6d86c // indirect
	github.com/iotaledger/hive.go/stringify v0.0.0-20230216132042-9c5c69b6d86c // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/net v0.0.0-20210410081132-afb366fc7cd1 // indirect
	golang.org/x/sys v0.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
