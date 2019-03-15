package main

import (
	"fmt"
	"os"

	"github.com/go-kit/kit/endpoint"
	cli "goa.design/plugins/goakit/examples/calc/gen/grpc/cli/calc"
	"google.golang.org/grpc"
)

func doGRPC(scheme, host string, timeout int, debug bool) (endpoint.Endpoint, interface{}, error) {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("could not connect to gRPC server at %s: %v", host, err))
	}
	return cli.ParseEndpoint(conn)
}
