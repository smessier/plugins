package goakit

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	goagrpc "goa.design/goa/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	// unaryHandler wraps the go-kit's gRPC server and implements goa gRPC's
	// UnaryHandler interface.
	unaryHandler struct {
		*kitgrpc.Server

		// response is the type returned by calling the go-kit endpoint. It is used
		// by the response encoder to build the protocol buffer message type.
		response interface{}
		// respmdfn is the function to encode gRPC header and trailer metadata.
		respmdfn goagrpc.ResponseEncoder
	}
)

// NewUnaryHandler returns a new goa handler for unary RPCs.
func NewUnaryHandler(e endpoint.Endpoint, dec kitgrpc.DecodeRequestFunc, enc kitgrpc.EncodeResponseFunc, respmdfn goagrpc.ResponseEncoder, options ...kitgrpc.ServerOption) goagrpc.UnaryHandler {
	if dec == nil {
		dec = func(context.Context, interface{}) (interface{}, error) { return nil, nil }
	}
	if enc == nil {
		enc = func(context.Context, interface{}) (interface{}, error) { return nil, nil }
	}
	s := unaryHandler{respmdfn: respmdfn}
	s.Server = kitgrpc.NewServer(s.wrap(e), dec, enc, options...)
	return &s
}

// Handle calls the ServeGRPC function to handle the RPC.
func (s *unaryHandler) Handle(ctx context.Context, reqpb interface{}) (respb interface{}, err error) {
	// call go-kit's ServeGRPC implementation
	_, respb, err = s.ServeGRPC(ctx, reqpb)
	if err != nil || s.respmdfn == nil || s.response == nil {
		return respb, err
	}

	// Encode header and trailer metadata
	hdr := metadata.MD{}
	trlr := metadata.MD{}
	if _, err = s.respmdfn(ctx, s.response, &hdr, &trlr); err != nil {
		return nil, err
	}
	if len(hdr) > 0 {
		if err := grpc.SendHeader(ctx, hdr); err != nil {
			return nil, err
		}
	}
	if len(trlr) > 0 {
		if err := grpc.SetTrailer(ctx, trlr); err != nil {
			return nil, err
		}
	}
	return respb, err
}

// wrap wraps go-kit endpoint to record the return type from invoking the
// endpoint. We do this because go-kit, at the time of writing, does not
// provide a way to populate the header and trailer metadata in the gRPC
// response using the go-kit endpoint return type. This is a fundamental
// requirement in goa as endpoint return types could be encoded in the response
// metadata.
func (s *unaryHandler) wrap(e endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		response, err = e(ctx, request)
		s.response = response
		return response, err
	}
}
