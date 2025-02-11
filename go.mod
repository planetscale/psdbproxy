module github.com/planetscale/psdbproxy

go 1.23.2

require (
	connectrpc.com/connect v1.14.0
	github.com/golang/glog v1.2.2
	github.com/planetscale/psdb v0.0.0-20231211201729-8cfd83fe2664
	github.com/planetscale/vitess-types v0.0.0-20231211191709-770e14433716
	github.com/spf13/pflag v1.0.5
	// Once Vitess v19 is out, we can move to that but for now
	// we depend on latest main here.
	// Once a fix for https://github.com/vitessio/vitess/issues/16015 is out, we can pin to that version.
	// For now, pin to pain with the fixes.
	vitess.io/vitess v0.10.3-0.20241106191105-e87457ee2401
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/klauspost/connect-compress/v2 v2.0.0 // indirect
	github.com/pires/go-proxyproto v0.7.0 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/slok/noglog v0.2.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.66.2 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/golang/glog => github.com/planetscale/noglog v0.2.1-0.20210421230640-bea75fcd2e8e
