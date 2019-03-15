package server

import (
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"goa.design/plugins/goakit/examples/calc/gen/grpc/calc/server"
	"google.golang.org/grpc/metadata"
)

func DecodeAddRequest() kitgrpc.DecodeRequestFunc {
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		return server.DecodeAddRequest(ctx, v, md)
	}
}

func EncodeAddResponse() kitgrpc.EncodeResponseFunc {
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		hdr := metadata.MD{}
		trlr := metadata.MD{}
		return server.EncodeAddResponse(ctx, v, &hdr, &trlr)
	}
}

func EncodeAddMetadata(ctx context.Context, v interface{}, hdr, trlr *metadata.MD) context.Context {
	return ctx
}
