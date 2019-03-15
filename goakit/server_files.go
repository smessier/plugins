package goakit

import (
	"fmt"
	"path"
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	grpccodegen "goa.design/goa/grpc/codegen"
)

// ServerFiles produces the files containing the gRPC server handler and init
// functions that configure the gRPC server to serve the requests.
func ServerFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, len(root.API.GRPC.Services))
	for i, svc := range root.API.GRPC.Services {
		fw[i] = serverFile(genpkg, svc)
	}
	return fw
}

// serverFile returns the file defining the mount handler functions for the given
// service.
func serverFile(genpkg string, svc *expr.GRPCServiceExpr) *codegen.File {
	data := grpccodegen.GRPCServices.Get(svc.Name())
	svcName := codegen.SnakeCase(data.Service.VarName)
	fpath := filepath.Join(codegen.Gendir, "grpc", svcName, "kitserver", "server.go")
	title := fmt.Sprintf("%s go-kit gRPC server", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "goa.design/plugins/goakit"},
			{Path: "github.com/go-kit/kit/endpoint"},
			{Path: "github.com/go-kit/kit/transport/grpc", Name: "kitgrpc"},
			{Path: "goa.design/goa/grpc", Name: "goagrpc"},
			{Path: path.Join(genpkg, svcName), Name: data.Service.PkgName},
			{Path: path.Join(genpkg, "grpc", svcName, "server")},
		}),
	}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "goakit-grpc-server-init",
		Source: grpcServerInitT,
		Data:   data,
	})
	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "goakit-grpc-handler-init",
			Source: grpcHandlerInitT,
			Data:   e,
		})
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

// input: grpccodegen.ServiceData
const grpcServerInitT = `{{ printf "%s instantiates the server struct with the %s service endpoints." .ServerInit .Service.Name | comment }}
func {{ .ServerInit }}(e *{{ .Service.PkgName }}.Endpoints{{ if .HasUnaryEndpoint }}, uh goagrpc.UnaryHandler{{ end }}{{ if .HasStreamingEndpoint }}, sh goagrpc.StreamHandler{{ end }}) *server.{{ .ServerStruct }} {
	return &server.{{ .ServerStruct }}{
	{{- range .Endpoints }}
		{{ .Method.VarName }}H: New{{ .Method.VarName }}Handler(e.{{ .Method.VarName }}{{ if .ServerStream }}, sh{{ else }}, uh{{ end }}),
	{{- end }}
	}
}
`

// input: grpccodegen.EndpointData
const grpcHandlerInitT = `{{ printf "New%sHandler creates a gRPC handler which serves the %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func New{{ .Method.VarName }}Handler(endpoint endpoint.Endpoint, h goagrpc.{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler, options ...kitgrpc.ServerOption) goagrpc.{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler {
	if h == nil {
		h = goakit.New{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler(endpoint, {{ if .Method.Payload }}Decode{{ .Method.VarName }}Request(){{ else }}nil{{ end }}{{ if not .ServerStream }}, Encode{{ .Method.VarName }}Response(), server.Encode{{ .Method.VarName }}Response{{ end }}, options...)
	}
	return h
}
`
