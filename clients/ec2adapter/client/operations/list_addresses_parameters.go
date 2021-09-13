// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewListAddressesParams creates a new ListAddressesParams object
// with the default values initialized.
func NewListAddressesParams() *ListAddressesParams {

	return &ListAddressesParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewListAddressesParamsWithTimeout creates a new ListAddressesParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewListAddressesParamsWithTimeout(timeout time.Duration) *ListAddressesParams {

	return &ListAddressesParams{

		timeout: timeout,
	}
}

// NewListAddressesParamsWithContext creates a new ListAddressesParams object
// with the default values initialized, and the ability to set a context for a request
func NewListAddressesParamsWithContext(ctx context.Context) *ListAddressesParams {

	return &ListAddressesParams{

		Context: ctx,
	}
}

// NewListAddressesParamsWithHTTPClient creates a new ListAddressesParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewListAddressesParamsWithHTTPClient(client *http.Client) *ListAddressesParams {

	return &ListAddressesParams{
		HTTPClient: client,
	}
}

/*ListAddressesParams contains all the parameters to send to the API endpoint
for the list addresses operation typically these are written to a http.Request
*/
type ListAddressesParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the list addresses params
func (o *ListAddressesParams) WithTimeout(timeout time.Duration) *ListAddressesParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list addresses params
func (o *ListAddressesParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list addresses params
func (o *ListAddressesParams) WithContext(ctx context.Context) *ListAddressesParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list addresses params
func (o *ListAddressesParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list addresses params
func (o *ListAddressesParams) WithHTTPClient(client *http.Client) *ListAddressesParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list addresses params
func (o *ListAddressesParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *ListAddressesParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}