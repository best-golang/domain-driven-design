// Code generated by go-swagger; DO NOT EDIT.

package product

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// FindAllHandlerFunc turns a function with the right signature into a find all handler
type FindAllHandlerFunc func(FindAllParams) middleware.Responder

// Handle executing the request and returning a response
func (fn FindAllHandlerFunc) Handle(params FindAllParams) middleware.Responder {
	return fn(params)
}

// FindAllHandler interface for that can handle valid find all params
type FindAllHandler interface {
	Handle(FindAllParams) middleware.Responder
}

// NewFindAll creates a new http.Handler for the find all operation
func NewFindAll(ctx *middleware.Context, handler FindAllHandler) *FindAll {
	return &FindAll{Context: ctx, Handler: handler}
}

/*FindAll swagger:route GET /product product findAll

FindAll find all API

*/
type FindAll struct {
	Context *middleware.Context
	Handler FindAllHandler
}

func (o *FindAll) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewFindAllParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
