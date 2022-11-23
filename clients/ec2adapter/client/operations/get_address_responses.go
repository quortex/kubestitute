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

// GetAddressReader is a Reader for the GetAddress structure.
type GetAddressReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetAddressReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetAddressOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetAddressBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewGetAddressForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetAddressNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetAddressInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetAddressOK creates a GetAddressOK with default headers values
func NewGetAddressOK() *GetAddressOK {
	return &GetAddressOK{}
}

/*
GetAddressOK handles this case with default header values.

OK
*/
type GetAddressOK struct {
	Payload *models.Address
}

func (o *GetAddressOK) Error() string {
	return fmt.Sprintf("[GET /addresses/{id}][%d] getAddressOK  %+v", 200, o.Payload)
}

func (o *GetAddressOK) GetPayload() *models.Address {
	return o.Payload
}

func (o *GetAddressOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Address)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAddressBadRequest creates a GetAddressBadRequest with default headers values
func NewGetAddressBadRequest() *GetAddressBadRequest {
	return &GetAddressBadRequest{}
}

/*
GetAddressBadRequest handles this case with default header values.

Bad Request
*/
type GetAddressBadRequest struct {
	Payload *models.ErrorResponse
}

func (o *GetAddressBadRequest) Error() string {
	return fmt.Sprintf("[GET /addresses/{id}][%d] getAddressBadRequest  %+v", 400, o.Payload)
}

func (o *GetAddressBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetAddressBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAddressForbidden creates a GetAddressForbidden with default headers values
func NewGetAddressForbidden() *GetAddressForbidden {
	return &GetAddressForbidden{}
}

/*
GetAddressForbidden handles this case with default header values.

Forbidden
*/
type GetAddressForbidden struct {
	Payload *models.ErrorResponse
}

func (o *GetAddressForbidden) Error() string {
	return fmt.Sprintf("[GET /addresses/{id}][%d] getAddressForbidden  %+v", 403, o.Payload)
}

func (o *GetAddressForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetAddressForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAddressNotFound creates a GetAddressNotFound with default headers values
func NewGetAddressNotFound() *GetAddressNotFound {
	return &GetAddressNotFound{}
}

/*
GetAddressNotFound handles this case with default header values.

Not Found
*/
type GetAddressNotFound struct {
	Payload *models.ErrorResponse
}

func (o *GetAddressNotFound) Error() string {
	return fmt.Sprintf("[GET /addresses/{id}][%d] getAddressNotFound  %+v", 404, o.Payload)
}

func (o *GetAddressNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetAddressNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAddressInternalServerError creates a GetAddressInternalServerError with default headers values
func NewGetAddressInternalServerError() *GetAddressInternalServerError {
	return &GetAddressInternalServerError{}
}

/*
GetAddressInternalServerError handles this case with default header values.

Internal Server Error
*/
type GetAddressInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *GetAddressInternalServerError) Error() string {
	return fmt.Sprintf("[GET /addresses/{id}][%d] getAddressInternalServerError  %+v", 500, o.Payload)
}

func (o *GetAddressInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetAddressInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
