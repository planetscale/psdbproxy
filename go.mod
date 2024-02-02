module github.com/planetscale/psdbproxy

go 1.21

require (
	connectrpc.com/connect v1.14.0
	github.com/golang/glog v1.2.0
	github.com/planetscale/psdb v0.0.0-20231211201729-8cfd83fe2664
	github.com/planetscale/vitess-types v0.0.0-20231211191709-770e14433716
	github.com/spf13/pflag v1.0.5
	// Once Vitess v19 is out, we can move to that but for now
	// we depend on latest main here.
	vitess.io/vitess v0.10.3-0.20240202095444-ea8a90d0938b
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/klauspost/compress v1.17.5 // indirect
	github.com/klauspost/connect-compress/v2 v2.0.0 // indirect
	github.com/pires/go-proxyproto v0.7.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/slok/noglog v0.2.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240125205218-1f4bbc51befe // indirect
	google.golang.org/grpc v1.61.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)

replace github.com/golang/glog => github.com/planetscale/noglog v0.2.1-0.20210421230640-bea75fcd2e8e
