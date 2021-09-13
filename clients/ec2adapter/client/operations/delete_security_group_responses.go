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

// DeleteSecurityGroupReader is a Reader for the DeleteSecurityGroup structure.
type DeleteSecurityGroupReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteSecurityGroupReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 204:
		result := NewDeleteSecurityGroupNoContent()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewDeleteSecurityGroupBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewDeleteSecurityGroupForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewDeleteSecurityGroupNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewDeleteSecurityGroupInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewDeleteSecurityGroupNoContent creates a DeleteSecurityGroupNoContent with default headers values
func NewDeleteSecurityGroupNoContent() *DeleteSecurityGroupNoContent {
	return &DeleteSecurityGroupNoContent{}
}

/*DeleteSecurityGroupNoContent handles this case with default header values.

No Content
*/
type DeleteSecurityGroupNoContent struct {
	Payload string
}

func (o *DeleteSecurityGroupNoContent) Error() string {
	return fmt.Sprintf("[DELETE /security-groups/{id}][%d] deleteSecurityGroupNoContent  %+v", 204, o.Payload)
}

func (o *DeleteSecurityGroupNoContent) GetPayload() string {
	return o.Payload
}

func (o *DeleteSecurityGroupNoContent) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteSecurityGroupBadRequest creates a DeleteSecurityGroupBadRequest with default headers values
func NewDeleteSecurityGroupBadRequest() *DeleteSecurityGroupBadRequest {
	return &DeleteSecurityGroupBadRequest{}
}

/*DeleteSecurityGroupBadRequest handles this case with default header values.

Bad Request
*/
type DeleteSecurityGroupBadRequest struct {
	Payload *models.ErrorResponse
}

func (o *DeleteSecurityGroupBadRequest) Error() string {
	return fmt.Sprintf("[DELETE /security-groups/{id}][%d] deleteSecurityGroupBadRequest  %+v", 400, o.Payload)
}

func (o *DeleteSecurityGroupBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteSecurityGroupBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteSecurityGroupForbidden creates a DeleteSecurityGroupForbidden with default headers values
func NewDeleteSecurityGroupForbidden() *DeleteSecurityGroupForbidden {
	return &DeleteSecurityGroupForbidden{}
}

/*DeleteSecurityGroupForbidden handles this case with default header values.

Forbidden
*/
type DeleteSecurityGroupForbidden struct {
	Payload *models.ErrorResponse
}

func (o *DeleteSecurityGroupForbidden) Error() string {
	return fmt.Sprintf("[DELETE /security-groups/{id}][%d] deleteSecurityGroupForbidden  %+v", 403, o.Payload)
}

func (o *DeleteSecurityGroupForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteSecurityGroupForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteSecurityGroupNotFound creates a DeleteSecurityGroupNotFound with default headers values
func NewDeleteSecurityGroupNotFound() *DeleteSecurityGroupNotFound {
	return &DeleteSecurityGroupNotFound{}
}

/*DeleteSecurityGroupNotFound handles this case with default header values.

Not Found
*/
type DeleteSecurityGroupNotFound struct {
	Payload *models.ErrorResponse
}

func (o *DeleteSecurityGroupNotFound) Error() string {
	return fmt.Sprintf("[DELETE /security-groups/{id}][%d] deleteSecurityGroupNotFound  %+v", 404, o.Payload)
}

func (o *DeleteSecurityGroupNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteSecurityGroupNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteSecurityGroupInternalServerError creates a DeleteSecurityGroupInternalServerError with default headers values
func NewDeleteSecurityGroupInternalServerError() *DeleteSecurityGroupInternalServerError {
	return &DeleteSecurityGroupInternalServerError{}
}

/*DeleteSecurityGroupInternalServerError handles this case with default header values.

Internal Server Error
*/
type DeleteSecurityGroupInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *DeleteSecurityGroupInternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /security-groups/{id}][%d] deleteSecurityGroupInternalServerError  %+v", 500, o.Payload)
}

func (o *DeleteSecurityGroupInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteSecurityGroupInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}