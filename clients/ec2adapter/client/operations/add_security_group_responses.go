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

// AddSecurityGroupReader is a Reader for the AddSecurityGroup structure.
type AddSecurityGroupReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *AddSecurityGroupReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewAddSecurityGroupCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewAddSecurityGroupBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewAddSecurityGroupForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewAddSecurityGroupInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewAddSecurityGroupCreated creates a AddSecurityGroupCreated with default headers values
func NewAddSecurityGroupCreated() *AddSecurityGroupCreated {
	return &AddSecurityGroupCreated{}
}

/*AddSecurityGroupCreated handles this case with default header values.

Created
*/
type AddSecurityGroupCreated struct {
	Payload string
}

func (o *AddSecurityGroupCreated) Error() string {
	return fmt.Sprintf("[POST /security-groups/{id}/associate][%d] addSecurityGroupCreated  %+v", 201, o.Payload)
}

func (o *AddSecurityGroupCreated) GetPayload() string {
	return o.Payload
}

func (o *AddSecurityGroupCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddSecurityGroupBadRequest creates a AddSecurityGroupBadRequest with default headers values
func NewAddSecurityGroupBadRequest() *AddSecurityGroupBadRequest {
	return &AddSecurityGroupBadRequest{}
}

/*AddSecurityGroupBadRequest handles this case with default header values.

Bad Request
*/
type AddSecurityGroupBadRequest struct {
	Payload *models.ErrorResponse
}

func (o *AddSecurityGroupBadRequest) Error() string {
	return fmt.Sprintf("[POST /security-groups/{id}/associate][%d] addSecurityGroupBadRequest  %+v", 400, o.Payload)
}

func (o *AddSecurityGroupBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *AddSecurityGroupBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddSecurityGroupForbidden creates a AddSecurityGroupForbidden with default headers values
func NewAddSecurityGroupForbidden() *AddSecurityGroupForbidden {
	return &AddSecurityGroupForbidden{}
}

/*AddSecurityGroupForbidden handles this case with default header values.

Forbidden
*/
type AddSecurityGroupForbidden struct {
	Payload *models.ErrorResponse
}

func (o *AddSecurityGroupForbidden) Error() string {
	return fmt.Sprintf("[POST /security-groups/{id}/associate][%d] addSecurityGroupForbidden  %+v", 403, o.Payload)
}

func (o *AddSecurityGroupForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *AddSecurityGroupForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddSecurityGroupInternalServerError creates a AddSecurityGroupInternalServerError with default headers values
func NewAddSecurityGroupInternalServerError() *AddSecurityGroupInternalServerError {
	return &AddSecurityGroupInternalServerError{}
}

/*AddSecurityGroupInternalServerError handles this case with default header values.

Internal Server Error
*/
type AddSecurityGroupInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *AddSecurityGroupInternalServerError) Error() string {
	return fmt.Sprintf("[POST /security-groups/{id}/associate][%d] addSecurityGroupInternalServerError  %+v", 500, o.Payload)
}

func (o *AddSecurityGroupInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *AddSecurityGroupInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
