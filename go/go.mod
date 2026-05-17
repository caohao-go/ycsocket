module server_golang

go 1.24.0

toolchain go1.24.4

replace github.com/mitchellh/mapstructure v1.5.0 => github.com/caohao-go/mapstructure v0.0.0-20250610075846-ab627e2e0496

require (
	git.code.oa.com/pcg-csd/trpc-ext/orm v1.14.6
	git.code.oa.com/pcg-csd/trpc-ext/redis v1.5.0
	git.code.oa.com/pcg-csd/trpc-ext/redis/trpc v1.5.0
	git.code.oa.com/pcg-csd/trpc-ext/util v1.4.3
	git.code.oa.com/trpc-go/trpc-database/timer v0.2.0
	git.code.oa.com/trpc-go/trpc-go v0.16.1
	git.woa.com/polaris/polaris-go/v2 v2.6.9
	git.woa.com/trpc-go/tnet/extensions/websocket v0.0.11
	git.woa.com/trpc-go/trpc-tnet-transport/websocket v0.1.10
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/dgraph-io/ristretto/v2 v2.4.0
	github.com/json-iterator/go v1.1.12
	github.com/martinlindhe/base36 v1.1.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/shopspring/decimal v1.4.0
	github.com/spf13/cast v1.7.0
	google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v3 v3.0.1
)

require (
	git.code.oa.com/trpc-go/trpc-database/mysql v0.3.1 // indirect
	git.code.oa.com/trpc-go/trpc-database/redis v0.2.3 // indirect
	git.code.oa.com/trpc-go/trpc-metrics-runtime v0.3.3 // indirect
	git.code.oa.com/trpc-go/trpc-selector-dsn v0.2.1 // indirect
	git.code.oa.com/trpc-go/trpc-utils v0.1.0 // indirect
	git.woa.com/jce/jce v1.2.0 // indirect
	git.woa.com/polaris/polaris-server-api/api/v1/model v1.1.3 // indirect
	git.woa.com/trpc-go/go_reuseport v1.7.0 // indirect
	git.woa.com/trpc-go/tnet v0.0.20 // indirect
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-redsync/redsync/v4 v4.5.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.1.0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gomodule/redigo v1.8.9 // indirect
	github.com/google/flatbuffers v2.0.0+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/panjf2000/ants/v2 v2.4.6 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.43.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/automaxprocs v1.3.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace git.woa.com/trpc-go/tnet/extensions/websocket v0.0.11 => ./local_mods/websocket
