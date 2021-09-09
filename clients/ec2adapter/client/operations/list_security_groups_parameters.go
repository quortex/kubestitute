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

// NewListSecurityGroupsParams creates a new ListSecurityGroupsParams object
// with the default values initialized.
func NewListSecurityGroupsParams() *ListSecurityGroupsParams {

	return &ListSecurityGroupsParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewListSecurityGroupsParamsWithTimeout creates a new ListSecurityGroupsParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewListSecurityGroupsParamsWithTimeout(timeout time.Duration) *ListSecurityGroupsParams {

	return &ListSecurityGroupsParams{

		timeout: timeout,
	}
}

// NewListSecurityGroupsParamsWithContext creates a new ListSecurityGroupsParams object
// with the default values initialized, and the ability to set a context for a request
func NewListSecurityGroupsParamsWithContext(ctx context.Context) *ListSecurityGroupsParams {

	return &ListSecurityGroupsParams{

		Context: ctx,
	}
}

// NewListSecurityGroupsParamsWithHTTPClient creates a new ListSecurityGroupsParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewListSecurityGroupsParamsWithHTTPClient(client *http.Client) *ListSecurityGroupsParams {

	return &ListSecurityGroupsParams{
		HTTPClient: client,
	}
}

/*ListSecurityGroupsParams contains all the parameters to send to the API endpoint
for the list security groups operation typically these are written to a http.Request
*/
type ListSecurityGroupsParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the list security groups params
func (o *ListSecurityGroupsParams) WithTimeout(timeout time.Duration) *ListSecurityGroupsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list security groups params
func (o *ListSecurityGroupsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list security groups params
func (o *ListSecurityGroupsParams) WithContext(ctx context.Context) *ListSecurityGroupsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list security groups params
func (o *ListSecurityGroupsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list security groups params
func (o *ListSecurityGroupsParams) WithHTTPClient(client *http.Client) *ListSecurityGroupsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list security groups params
func (o *ListSecurityGroupsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *ListSecurityGroupsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
