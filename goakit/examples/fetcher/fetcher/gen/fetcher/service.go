// Code generated by goa v3.0.0, DO NOT EDIT.
//
// fetcher service
//
// Command:
// $ goa gen goa.design/plugins/goakit/examples/fetcher/fetcher/design -o
// $(GOPATH)/src/goa.design/plugins/goakit/examples/fetcher/fetcher

package fetcher

import (
	"context"

	goa "goa.design/goa/v3"
	fetcherviews "goa.design/plugins/goakit/examples/fetcher/fetcher/gen/fetcher/views"
)

// Service is the fetcher service interface.
type Service interface {
	// Fetch makes a GET request to the given URL and stores the results in the
	// archiver service which must be running or the request fails
	Fetch(context.Context, *FetchPayload) (res *FetchMedia, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "fetcher"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"fetch"}

// FetchPayload is the payload type of the fetcher service fetch method.
type FetchPayload struct {
	// URL to be fetched
	URL string
}

// FetchMedia is the result type of the fetcher service fetch method.
type FetchMedia struct {
	// HTTP status code returned by fetched service
	Status int
	// The href to the corresponding archive in the archiver service
	ArchiveHref string
}

// MakeBadRequest builds a goa.ServiceError from an error.
func MakeBadRequest(err error) *goa.ServiceError {
	return &goa.ServiceError{
		Name:    "bad_request",
		ID:      goa.NewErrorID(),
		Message: err.Error(),
	}
}

// MakeInternalError builds a goa.ServiceError from an error.
func MakeInternalError(err error) *goa.ServiceError {
	return &goa.ServiceError{
		Name:    "internal_error",
		ID:      goa.NewErrorID(),
		Message: err.Error(),
	}
}

// NewFetchMedia initializes result type FetchMedia from viewed result type
// FetchMedia.
func NewFetchMedia(vres *fetcherviews.FetchMedia) *FetchMedia {
	var res *FetchMedia
	switch vres.View {
	case "default", "":
		res = newFetchMedia(vres.Projected)
	}
	return res
}

// NewViewedFetchMedia initializes viewed result type FetchMedia from result
// type FetchMedia using the given view.
func NewViewedFetchMedia(res *FetchMedia, view string) *fetcherviews.FetchMedia {
	var vres *fetcherviews.FetchMedia
	switch view {
	case "default", "":
		p := newFetchMediaView(res)
		vres = &fetcherviews.FetchMedia{p, "default"}
	}
	return vres
}

// newFetchMedia converts projected type FetchMedia to service type FetchMedia.
func newFetchMedia(vres *fetcherviews.FetchMediaView) *FetchMedia {
	res := &FetchMedia{}
	if vres.Status != nil {
		res.Status = *vres.Status
	}
	if vres.ArchiveHref != nil {
		res.ArchiveHref = *vres.ArchiveHref
	}
	return res
}

// newFetchMediaView projects result type FetchMedia into projected type
// FetchMediaView using the "default" view.
func newFetchMediaView(res *FetchMedia) *fetcherviews.FetchMediaView {
	vres := &fetcherviews.FetchMediaView{
		Status:      &res.Status,
		ArchiveHref: &res.ArchiveHref,
	}
	return vres
}
