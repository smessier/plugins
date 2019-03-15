package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcmdlwr "goa.design/goa/grpc/middleware"
	"goa.design/goa/middleware"
	"goa.design/plugins/goakit"
	calcsvc "goa.design/plugins/goakit/examples/calc/gen/calc"
	calcsvckitsvr "goa.design/plugins/goakit/examples/calc/gen/grpc/calc/kitserver"
	calcpb "goa.design/plugins/goakit/examples/calc/gen/grpc/calc/pb"
	calcsvcsvr "goa.design/plugins/goakit/examples/calc/gen/grpc/calc/server"
	"google.golang.org/grpc"
)

// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, calcEndpoints *calcsvc.Endpoints, wg *sync.WaitGroup, errc chan error, logger log.Logger, debug bool) {
	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
		//calcServer    *calcsvcsvr.Server
		calcKitServer *goakit.GRPCServer
	)
	{
		//calcServer = calcsvcsvr.New(calcEndpoints, nil)
		calcKitServer = goakit.NewGRPCServer(
			endpoint.Endpoint(calcEndpoints.Add),
			calcsvckitsvr.DecodeAddRequest,
			calcsvckitsvr.EncodeAddResponse,
			calcsvckitsvr.EncodeAddMetadata,
		)
	}

	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			grpcmdlwr.UnaryRequestID(),
			grpcmdlwr.UnaryServerLog(adapter),
		),
	)

	// Register the servers.
	//calcpb.RegisterCalcServer(srv, calcServer)
	calcpb.RegisterCalcServer(srv, calcKitServer)

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
