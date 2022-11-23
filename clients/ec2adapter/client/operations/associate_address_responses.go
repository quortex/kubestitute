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

// AssociateAddressReader is a Reader for the AssociateAddress structure.
type AssociateAddressReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *AssociateAddressReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewAssociateAddressCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewAssociateAddressBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewAssociateAddressForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewAssociateAddressNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewAssociateAddressInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewAssociateAddressCreated creates a AssociateAddressCreated with default headers values
func NewAssociateAddressCreated() *AssociateAddressCreated {
	return &AssociateAddressCreated{}
}

/*
AssociateAddressCreated handles this case with default header values.

Created
*/
type AssociateAddressCreated struct {
	Payload string
}

func (o *AssociateAddressCreated) Error() string {
	return fmt.Sprintf("[POST /addresses/{id}/associate][%d] associateAddressCreated  %+v", 201, o.Payload)
}

func (o *AssociateAddressCreated) GetPayload() string {
	return o.Payload
}

func (o *AssociateAddressCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAssociateAddressBadRequest creates a AssociateAddressBadRequest with default headers values
func NewAssociateAddressBadRequest() *AssociateAddressBadRequest {
	return &AssociateAddressBadRequest{}
}

/*
AssociateAddressBadRequest handles this case with default header values.

Bad Request
*/
type AssociateAddressBadRequest struct {
	Payload *models.ErrorResponse
}

func (o *AssociateAddressBadRequest) Error() string {
	return fmt.Sprintf("[POST /addresses/{id}/associate][%d] associateAddressBadRequest  %+v", 400, o.Payload)
}

func (o *AssociateAddressBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *AssociateAddressBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAssociateAddressForbidden creates a AssociateAddressForbidden with default headers values
func NewAssociateAddressForbidden() *AssociateAddressForbidden {
	return &AssociateAddressForbidden{}
}

/*
AssociateAddressForbidden handles this case with default header values.

Forbidden
*/
type AssociateAddressForbidden struct {
	Payload *models.ErrorResponse
}

func (o *AssociateAddressForbidden) Error() string {
	return fmt.Sprintf("[POST /addresses/{id}/associate][%d] associateAddressForbidden  %+v", 403, o.Payload)
}

func (o *AssociateAddressForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *AssociateAddressForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAssociateAddressNotFound creates a AssociateAddressNotFound with default headers values
func NewAssociateAddressNotFound() *AssociateAddressNotFound {
	return &AssociateAddressNotFound{}
}

/*
AssociateAddressNotFound handles this case with default header values.

Not Found
*/
type AssociateAddressNotFound struct {
	Payload *models.ErrorResponse
}

func (o *AssociateAddressNotFound) Error() string {
	return fmt.Sprintf("[POST /addresses/{id}/associate][%d] associateAddressNotFound  %+v", 404, o.Payload)
}

func (o *AssociateAddressNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *AssociateAddressNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAssociateAddressInternalServerError creates a AssociateAddressInternalServerError with default headers values
func NewAssociateAddressInternalServerError() *AssociateAddressInternalServerError {
	return &AssociateAddressInternalServerError{}
}

/*
AssociateAddressInternalServerError handles this case with default header values.

Internal Server Error
*/
type AssociateAddressInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *AssociateAddressInternalServerError) Error() string {
	return fmt.Sprintf("[POST /addresses/{id}/associate][%d] associateAddressInternalServerError  %+v", 500, o.Payload)
}

func (o *AssociateAddressInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *AssociateAddressInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
