package auth

import (
	"net/http"
	"strings"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type guestAuthRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type GuestAuthForm struct {
	DeviceID   string
	DeviceName string
}

func NewGuestAuthForm() *GuestAuthForm {
	return &GuestAuthForm{}
}

func (f *GuestAuthForm) ParseAndValidate(r *http.Request) (*GuestAuthForm, response.Error) {
	var request *guestAuthRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetDeviceID(request, errors)
	f.validateAndSetDeviceName(request)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *GuestAuthForm) validateAndSetDeviceID(req *guestAuthRequest, errors map[string]response.ErrorMessage) {
	deviceID := strings.TrimSpace(req.DeviceID)
	if deviceID == "" {
		errors["device_id"] = response.MissedValueError("missed value")

		return
	}

	f.DeviceID = deviceID
}

func (f *GuestAuthForm) validateAndSetDeviceName(request *guestAuthRequest) {
	deviceName := strings.TrimSpace(request.DeviceName)
	if deviceName == "" {
		return
	}

	f.DeviceName = deviceName
}
