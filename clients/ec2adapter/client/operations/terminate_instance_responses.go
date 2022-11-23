// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"quortex.io/kubestitute/clients/ec2adapter/models"
)

// TerminateInstanceReader is a Reader for the TerminateInstance structure.
type TerminateInstanceReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *TerminateInstanceReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 204:
		result := NewTerminateInstanceNoContent()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewTerminateInstanceBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewTerminateInstanceForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewTerminateInstanceNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewTerminateInstanceInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewTerminateInstanceNoContent creates a TerminateInstanceNoContent with default headers values
func NewTerminateInstanceNoContent() *TerminateInstanceNoContent {
	return &TerminateInstanceNoContent{}
}

/*
TerminateInstanceNoContent handles this case with default header values.

No Content
*/
type TerminateInstanceNoContent struct {
	Payload string
}

func (o *TerminateInstanceNoContent) Error() string {
	return fmt.Sprintf("[DELETE /instances/{id}][%d] terminateInstanceNoContent  %+v", 204, o.Payload)
}

func (o *TerminateInstanceNoContent) GetPayload() string {
	return o.Payload
}

func (o *TerminateInstanceNoContent) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewTerminateInstanceBadRequest creates a TerminateInstanceBadRequest with default headers values
func NewTerminateInstanceBadRequest() *TerminateInstanceBadRequest {
	return &TerminateInstanceBadRequest{}
}

/*
TerminateInstanceBadRequest handles this case with default header values.

Bad Request
*/
type TerminateInstanceBadRequest struct {
	Payload *models.ErrorResponse
}

func (o *TerminateInstanceBadRequest) Error() string {
	return fmt.Sprintf("[DELETE /instances/{id}][%d] terminateInstanceBadRequest  %+v", 400, o.Payload)
}

func (o *TerminateInstanceBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *TerminateInstanceBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewTerminateInstanceForbidden creates a TerminateInstanceForbidden with default headers values
func NewTerminateInstanceForbidden() *TerminateInstanceForbidden {
	return &TerminateInstanceForbidden{}
}

/*
TerminateInstanceForbidden handles this case with default header values.

Forbidden
*/
type TerminateInstanceForbidden struct {
	Payload *models.ErrorResponse
}

func (o *TerminateInstanceForbidden) Error() string {
	return fmt.Sprintf("[DELETE /instances/{id}][%d] terminateInstanceForbidden  %+v", 403, o.Payload)
}

func (o *TerminateInstanceForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *TerminateInstanceForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewTerminateInstanceNotFound creates a TerminateInstanceNotFound with default headers values
func NewTerminateInstanceNotFound() *TerminateInstanceNotFound {
	return &TerminateInstanceNotFound{}
}

/*
TerminateInstanceNotFound handles this case with default header values.

Not Found
*/
type TerminateInstanceNotFound struct {
	Payload *models.ErrorResponse
}

func (o *TerminateInstanceNotFound) Error() string {
	return fmt.Sprintf("[DELETE /instances/{id}][%d] terminateInstanceNotFound  %+v", 404, o.Payload)
}

func (o *TerminateInstanceNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *TerminateInstanceNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewTerminateInstanceInternalServerError creates a TerminateInstanceInternalServerError with default headers values
func NewTerminateInstanceInternalServerError() *TerminateInstanceInternalServerError {
	return &TerminateInstanceInternalServerError{}
}

/*
TerminateInstanceInternalServerError handles this case with default header values.

Internal Server Error
*/
type TerminateInstanceInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *TerminateInstanceInternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /instances/{id}][%d] terminateInstanceInternalServerError  %+v", 500, o.Payload)
}

func (o *TerminateInstanceInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *TerminateInstanceInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
