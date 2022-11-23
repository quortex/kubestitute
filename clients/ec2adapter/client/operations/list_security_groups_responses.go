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

// ListSecurityGroupsReader is a Reader for the ListSecurityGroups structure.
type ListSecurityGroupsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListSecurityGroupsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListSecurityGroupsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewListSecurityGroupsBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewListSecurityGroupsForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewListSecurityGroupsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListSecurityGroupsOK creates a ListSecurityGroupsOK with default headers values
func NewListSecurityGroupsOK() *ListSecurityGroupsOK {
	return &ListSecurityGroupsOK{}
}

/*
ListSecurityGroupsOK handles this case with default header values.

OK
*/
type ListSecurityGroupsOK struct {
	Payload []*models.SecurityGroup
}

func (o *ListSecurityGroupsOK) Error() string {
	return fmt.Sprintf("[GET /security-groups][%d] listSecurityGroupsOK  %+v", 200, o.Payload)
}

func (o *ListSecurityGroupsOK) GetPayload() []*models.SecurityGroup {
	return o.Payload
}

func (o *ListSecurityGroupsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListSecurityGroupsBadRequest creates a ListSecurityGroupsBadRequest with default headers values
func NewListSecurityGroupsBadRequest() *ListSecurityGroupsBadRequest {
	return &ListSecurityGroupsBadRequest{}
}

/*
ListSecurityGroupsBadRequest handles this case with default header values.

Bad Request
*/
type ListSecurityGroupsBadRequest struct {
	Payload *models.ErrorResponse
}

func (o *ListSecurityGroupsBadRequest) Error() string {
	return fmt.Sprintf("[GET /security-groups][%d] listSecurityGroupsBadRequest  %+v", 400, o.Payload)
}

func (o *ListSecurityGroupsBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *ListSecurityGroupsBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListSecurityGroupsForbidden creates a ListSecurityGroupsForbidden with default headers values
func NewListSecurityGroupsForbidden() *ListSecurityGroupsForbidden {
	return &ListSecurityGroupsForbidden{}
}

/*
ListSecurityGroupsForbidden handles this case with default header values.

Forbidden
*/
type ListSecurityGroupsForbidden struct {
	Payload *models.ErrorResponse
}

func (o *ListSecurityGroupsForbidden) Error() string {
	return fmt.Sprintf("[GET /security-groups][%d] listSecurityGroupsForbidden  %+v", 403, o.Payload)
}

func (o *ListSecurityGroupsForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *ListSecurityGroupsForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListSecurityGroupsInternalServerError creates a ListSecurityGroupsInternalServerError with default headers values
func NewListSecurityGroupsInternalServerError() *ListSecurityGroupsInternalServerError {
	return &ListSecurityGroupsInternalServerError{}
}

/*
ListSecurityGroupsInternalServerError handles this case with default header values.

Internal Server Error
*/
type ListSecurityGroupsInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *ListSecurityGroupsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /security-groups][%d] listSecurityGroupsInternalServerError  %+v", 500, o.Payload)
}

func (o *ListSecurityGroupsInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *ListSecurityGroupsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
