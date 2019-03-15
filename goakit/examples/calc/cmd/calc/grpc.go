package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"

	"github.com/go-kit/kit/log"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcmdlwr "goa.design/goa/grpc/middleware"
	calc "goa.design/plugins/goakit/examples/calc/gen/calc"
	calckitsvr "goa.design/plugins/goakit/examples/calc/gen/grpc/calc/kitserver"
	calcpb "goa.design/plugins/goakit/examples/calc/gen/grpc/calc/pb"
	calcsvr "goa.design/plugins/goakit/examples/calc/gen/grpc/calc/server"
	"google.golang.org/grpc"
)

// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, calcEndpoints *calc.Endpoints, wg *sync.WaitGroup, errc chan error, logger log.Logger, debug bool) {

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
		calcServer *calcsvr.Server
	)
	{
		calcServer = calckitsvr.New(calcEndpoints, nil)
	}

	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			grpcmdlwr.UnaryRequestID(),
			grpcmdlwr.UnaryServerLog(logger),
		),
	)

	// Register the servers.
	calcpb.RegisterCalcServer(srv, calcServer)

	for svc, info := range srv.GetServiceInfo() {
		for _, m := range info.Methods {
			logger.Log("info", fmt.Sprintf("serving gRPC method %s", svc+"/"+m.Name))
		}
	}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start gRPC server in a separate goroutine.
		go func() {
			lis, err := net.Listen("tcp", u.Host)
			if err != nil {
				errc <- err
			}
			logger.Log("info", fmt.Sprintf("gRPC server listening on %q", u.Host))
			errc <- srv.Serve(lis)
		}()

		<-ctx.Done()
		logger.Log("info", fmt.Sprintf("shutting down gRPC server at %q", u.Host))
		srv.Stop()
	}()
}
