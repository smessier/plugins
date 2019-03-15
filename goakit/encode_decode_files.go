package goakit

import (
	"fmt"
	"path"
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	grpccodegen "goa.design/goa/grpc/codegen"
	httpcodegen "goa.design/goa/http/codegen"
)

// EncodeDecodeFiles produces a set of go-kit transport encoders and decoders
// that wrap the corresponding generated goa functions.
func EncodeDecodeFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, r := range root.API.HTTP.Services {
		fw = append(fw, httpServerEncodeDecode(genpkg, r))
		fw = append(fw, httpClientEncodeDecode(genpkg, r))
	}
	for _, r := range root.API.GRPC.Services {
		fw = append(fw, grpcServerEncodeDecode(genpkg, r))
	}
	return fw
}

// httpServerEncodeDecode returns the file defining the go-kit HTTP server encoding
// and decoding logic.
func httpServerEncodeDecode(genpkg string, svc *expr.HTTPServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "kitserver", "encode_decode.go")
	data := httpcodegen.HTTPServices.Get(svc.Name())
	title := fmt.Sprintf("%s go-kit HTTP server encoders and decoders", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "net/http"},
			{Path: "strings"},
			{Path: "github.com/go-kit/kit/transport/http", Name: "kithttp"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: "goa.design/goa/http", Name: "goahttp"},
			{Path: genpkg + "/http/" + data.Service.Name + "/server"},
		}),
	}

	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "goakit-response-encoder",
			Source: httpResponseEncoderT,
			Data:   e,
		})

		if e.Payload.Ref != "" {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "goakit-request-decoder",
				Source: httpRequestDecoderT,
				Data:   e,
			})
		}

		if len(e.Errors) > 0 {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "goakit-error-encoder",
				Source: errorEncoderT,
				Data:   e,
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// httpClientEncodeDecode returns the file defining the go-kit HTTP client encoding
// and decoding logic.
func httpClientEncodeDecode(genpkg string, svc *expr.HTTPServiceExpr) *codegen.File {
	data := httpcodegen.HTTPServices.Get(svc.Name())
	svcName := codegen.SnakeCase(data.Service.VarName)
	path := filepath.Join(codegen.Gendir, "http", svcName, "kitclient", "encode_decode.go")
	title := fmt.Sprintf("%s go-kit HTTP client encoders and decoders", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "net/http"},
			{Path: "strings"},
			{Path: "github.com/go-kit/kit/transport/http", Name: "kithttp"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: "goa.design/goa/http", Name: "goahttp"},
			{Path: genpkg + "/http/" + svcName + "/client"},
		}),
	}

	for _, e := range data.Endpoints {
		if e.RequestEncoder != "" {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "goakit-request-encoder",
				Source: httpRequestEncoderT,
				Data:   e,
			})
		}
		if e.Result != nil || len(e.Errors) > 0 {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "goakit-response-decoder",
				Source: httpResponseDecoderT,
				Data:   e,
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func grpcServerEncodeDecode(genpkg string, svc *expr.GRPCServiceExpr) *codegen.File {
	data := grpccodegen.GRPCServices.Get(svc.Name())
	svcName := codegen.SnakeCase(data.Service.VarName)
	fpath := filepath.Join(codegen.Gendir, "grpc", svcName, "kitserver", "encode_decode.go")
	title := fmt.Sprintf("%s go-kit gRPC server encoders and decoders", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "google.golang.org/grpc/metadata"},
			{Path: "github.com/go-kit/kit/transport/grpc", Name: "kitgrpc"},
			{Path: path.Join(genpkg, "grpc", svcName, "server")},
		}),
	}

	for _, e := range data.Endpoints {
		if e.PayloadRef != "" {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "goakit-request-decoder",
				Source: grpcRequestDecoderT,
				Data:   e,
			})
		}
		if e.ResultRef != "" {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "goakit-response-encoder",
				Source: grpcResponseEncoderT,
				Data:   e,
			})
		}
	}

	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

// input: EndpointData
const httpRequestEncoderT = `{{ printf "%s returns a go-kit EncodeRequestFunc suitable for encoding %s %s requests." .RequestEncoder .ServiceName .Method.Name | comment }}
func {{ .RequestEncoder }}(encoder func(*http.Request) goahttp.Encoder) kithttp.EncodeRequestFunc {
	enc := client.{{ .RequestEncoder }}(encoder)
	return func(_ context.Context, r *http.Request, v interface{}) error {
		return enc(r, v)
	}
}
`

// input: EndpointData
const httpRequestDecoderT = `{{ printf "%s returns a go-kit DecodeRequestFunc suitable for decoding %s %s requests." .RequestDecoder .ServiceName .Method.Name | comment }}
func {{ .RequestDecoder }}(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) kithttp.DecodeRequestFunc {
	dec := server.{{ .RequestDecoder }}(mux, decoder)
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		r = r.WithContext(ctx)
		return dec(r)
	}
}
`

// input: EndpointData
const httpResponseEncoderT = `{{ printf "%s returns a go-kit EncodeResponseFunc suitable for encoding %s %s responses." .ResponseEncoder .ServiceName .Method.Name | comment }}
 func {{ .ResponseEncoder }}(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) kithttp.EncodeResponseFunc {
 	return server.{{ .ResponseEncoder }}(encoder)
 }
`

// input: EndpointData
const errorEncoderT = `{{ printf "%s returns a go-kit EncodeResponseFunc suitable for encoding errors returned by the %s %s endpoint." .ErrorEncoder .ServiceName .Method.Name | comment }}
 func {{ .ErrorEncoder }}(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) kithttp.ErrorEncoder {
 	enc := server.{{ .ErrorEncoder }}(encoder)
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		enc(ctx, w, err)
	}
}
`

// input: EndpointData
const httpResponseDecoderT = `{{ printf "%s returns a go-kit DecodeResponseFunc suitable for decoding %s %s responses." .ResponseDecoder .ServiceName .Method.Name | comment }}
func {{ .ResponseDecoder }}(decoder func(*http.Response) goahttp.Decoder) kithttp.DecodeResponseFunc {
	dec := client.{{ .ResponseDecoder }}(decoder, false)
	return func(ctx context.Context, resp *http.Response) (interface{}, error) {
		return dec(resp)
	}
}
`

// input: grpccodegen.EndpointData
const grpcRequestDecoderT = `{{ printf "Decode%sRequest returns a go-kit DecodeRequestFunc suitable for decoding %s %s requests." .Method.VarName .ServiceName .Method.Name | comment }}
func Decode{{ .Method.VarName }}Request() kitgrpc.DecodeRequestFunc {
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		return server.Decode{{ .Method.VarName }}Request(ctx, v, md)
	}
}
`

// input: grpccodegen.EndpointData
const grpcResponseEncoderT = `{{ printf "Encode%sResponse returns a go-kit EncodeResponseFunc suitable for encoding %s %s responses." .Method.VarName .ServiceName .Method.Name | comment }}
func Encode{{ .Method.VarName }}Response() kitgrpc.EncodeResponseFunc {
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		md := metadata.MD{}
		return server.Encode{{ .Method.VarName }}Response(ctx, v, &md, &md)
	}
}
`
