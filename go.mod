module github.com/planetscale/psdbproxy

go 1.24.13

require (
	connectrpc.com/connect v1.18.1
	github.com/golang/glog v1.2.4
	github.com/planetscale/psdb v0.0.0-20260313223120-d44ec59fda55
	github.com/planetscale/vitess-types v0.0.0-20260313221731-c96dbf730f7d
	github.com/spf13/pflag v1.0.6
	vitess.io/vitess v0.22.4
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/connect-compress/v2 v2.0.0 // indirect
	github.com/pires/go-proxyproto v0.8.0 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20241121165744-79df5c4772f2 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/slok/noglog v0.2.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250313205543-e70fdf4c4cb4 // indirect
	google.golang.org/grpc v1.71.1 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace github.com/golang/glog => github.com/planetscale/noglog v0.2.1-0.20210421230640-bea75fcd2e8e
