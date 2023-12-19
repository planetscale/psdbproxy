module github.com/planetscale/psdbproxy

go 1.21

require (
	connectrpc.com/connect v1.12.0
	github.com/golang/glog v1.1.2
	github.com/planetscale/psdb v0.0.0-20231031205845-c667a0f67df1
	github.com/planetscale/vitess-types v0.0.0-20231005172453-a4a7c0c1df6a
	github.com/spf13/pflag v1.0.5
	// Once Vitess v19 is out, we can move to that but for now
	// we depend on latest main here.
	vitess.io/vitess v0.0.0-20231219153810-071454f8173b
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/klauspost/connect-compress/v2 v2.0.0 // indirect
	github.com/pires/go-proxyproto v0.7.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/slok/noglog v0.2.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/sync v0.4.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/golang/glog => github.com/planetscale/noglog v0.2.1-0.20210421230640-bea75fcd2e8e
