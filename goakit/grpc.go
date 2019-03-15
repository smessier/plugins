package goakit

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	// GRPCServer wraps the go-kit's gRPC server and implements the go-kit's gRPC
	// Handler interface.
	GRPCServer struct {
		*grpc.Server

		// response is the type returned by calling the go-kit endpoint.
		response interface{}
		// respfn is the function to encode gRPC header and trailer metadata.
		respfn EncodeResponseFunc
	}

	// EncodeResponseFunc manipulates the header and trailer metadata using the
	// request context and endpoint response.
	EncodeResponseFunc func(ctx context.Context, response interface{}, hdr, trlr *metadata.MD) context.Context
)

// NewGRPCServer returns a new goakit gRPC server.
func NewGRPCServer(e endpoint.Endpoint, dec grpc.DecodeRequestFunc, enc grpc.EncodeResponseFunc, respfn EncodeResponseFunc, options ...grpc.ServerOption) *GRPCServer {
	s := GRPCServer{respfn: respfn}
	s.Server = grpc.NewServer(s.wrap(e), dec, enc, options...)
	return &s
}

// ServeGRPC implements go-kit's Handler interface.
func (s *GRPCServer) ServeGRPC(ctx context.Context, request interface{}) (context.Context, interface{}, error) {
	// call go-kit's ServeGRPC implementation
	ctx, resp, err := s.Server.ServeGRPC(ctx, request)
	if err != nil || s.respfn == nil || s.response == nil {
		return ctx, resp, err
	}

	// Encode header and trailer metadata
	hdr := metadata.MD{}
	trlr := metadata.MD{}
	ctx = s.respfn(ctx, s.response, &hdr, &trlr)
	if len(hdr) > 0 {
		if err := grpc.SendHeader(ctx, hdr); err != nil {
			return ctx, nil, err
		}
	}
	if len(trlr) > 0 {
		if err := grpc.SetTrailer(ctx, trlr); err != nil {
			return ctx, nil, err
		}
	}
	return ctx, resp, nil
}

// Wrap go-kit endpoint to record the return type from invoking the endpoint.
// We do this because go-kit, at the time of writing, does not provide a way
// to populate the header and trailer metadata in the gRPC response using the
// return type obtained by invoking the go-kit endpoint. This is a
// fundamental requirement in goa as endpoint return types could be encoded in
// the response metadata if specified in the desgin.
func (s *GRPCServer) wrap(e endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		response, err := e(ctx, request)
		s.response = response
		return response, err
	}
}
