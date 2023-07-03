package auth

import (
	"net/http"
	"strings"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type byDeviceRequest struct {
	DeviceID string `json:"device_id"`
}

type ByDeviceForm struct {
	DeviceID string
}

func NewByDeviceForm() *ByDeviceForm {
	return &ByDeviceForm{}
}

func (f *ByDeviceForm) ParseAndValidate(r *http.Request) (*ByDeviceForm, response.Error) {
	var request *byDeviceRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetDeviceID(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *ByDeviceForm) validateAndSetDeviceID(req *byDeviceRequest, errors map[string]response.ErrorMessage) {
	deviceID := strings.TrimSpace(req.DeviceID)
	if deviceID == "" {
		errors["device_id"] = response.MissedValueError("missed value")

		return
	}

	f.DeviceID = deviceID
}
