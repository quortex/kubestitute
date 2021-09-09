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

	"quortex.io/kubestitute/clients/ec2adapter/models"
)

// NewDetachAutoscalingGroupInstancesParams creates a new DetachAutoscalingGroupInstancesParams object
// with the default values initialized.
func NewDetachAutoscalingGroupInstancesParams() *DetachAutoscalingGroupInstancesParams {
	var ()
	return &DetachAutoscalingGroupInstancesParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewDetachAutoscalingGroupInstancesParamsWithTimeout creates a new DetachAutoscalingGroupInstancesParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewDetachAutoscalingGroupInstancesParamsWithTimeout(timeout time.Duration) *DetachAutoscalingGroupInstancesParams {
	var ()
	return &DetachAutoscalingGroupInstancesParams{

		timeout: timeout,
	}
}

// NewDetachAutoscalingGroupInstancesParamsWithContext creates a new DetachAutoscalingGroupInstancesParams object
// with the default values initialized, and the ability to set a context for a request
func NewDetachAutoscalingGroupInstancesParamsWithContext(ctx context.Context) *DetachAutoscalingGroupInstancesParams {
	var ()
	return &DetachAutoscalingGroupInstancesParams{

		Context: ctx,
	}
}

// NewDetachAutoscalingGroupInstancesParamsWithHTTPClient creates a new DetachAutoscalingGroupInstancesParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewDetachAutoscalingGroupInstancesParamsWithHTTPClient(client *http.Client) *DetachAutoscalingGroupInstancesParams {
	var ()
	return &DetachAutoscalingGroupInstancesParams{
		HTTPClient: client,
	}
}

/*DetachAutoscalingGroupInstancesParams contains all the parameters to send to the API endpoint
for the detach autoscaling group instances operation typically these are written to a http.Request
*/
type DetachAutoscalingGroupInstancesParams struct {

	/*Name
	  The AutoscalingGroup name

	*/
	Name string
	/*Request
	  The AutoscalingGroup DetachInstances request

	*/
	Request *models.DetachInstancesRequest

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the detach autoscaling group instances params
func (o *DetachAutoscalingGroupInstancesParams) WithTimeout(timeout time.Duration) *DetachAutoscalingGroupInstancesParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the detach autoscaling group instances params
func (o *DetachAutoscalingGroupInstancesParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the detach autoscaling group instances params
func (o *DetachAutoscalingGroupInstancesParams) WithContext(ctx context.Context) *DetachAutoscalingGroupInstancesParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the detach autoscaling group instances params
func (o *DetachAutoscalingGroupInstancesParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the detach autoscaling group instances params
func (o *DetachAutoscalingGroupInstancesParams) WithHTTPClient(client *http.Client) *DetachAutoscalingGroupInstancesParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the detach autoscaling group instances params
func (o *DetachAutoscalingGroupInstancesParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithName adds the name to the detach autoscaling group instances params
func (o *DetachAutoscalingGroupInstancesParams) WithName(name string) *DetachAutoscalingGroupInstancesParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the detach autoscaling group instances params
func (o *DetachAutoscalingGroupInstancesParams) SetName(name string) {
	o.Name = name
}

// WithRequest adds the request to the detach autoscaling group instances params
func (o *DetachAutoscalingGroupInstancesParams) WithRequest(request *models.DetachInstancesRequest) *DetachAutoscalingGroupInstancesParams {
	o.SetRequest(request)
	return o
}

// SetRequest adds the request to the detach autoscaling group instances params
func (o *DetachAutoscalingGroupInstancesParams) SetRequest(request *models.DetachInstancesRequest) {
	o.Request = request
}

// WriteToRequest writes these params to a swagger request
func (o *DetachAutoscalingGroupInstancesParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param name
	if err := r.SetPathParam("name", o.Name); err != nil {
		return err
	}

	if o.Request != nil {
		if err := r.SetBodyParam(o.Request); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
