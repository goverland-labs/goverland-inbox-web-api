package auth

import (
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type siweAuthRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Message    string `json:"message"`
	Address    string `json:"address"`
	Signature  string `json:"signature"`
}

type SiweAuthForm struct {
	DeviceID   string
	DeviceName string
	Message    string
	Address    common.Address
	Signature  string
}

func NewSiweAuthForm() *SiweAuthForm {
	return &SiweAuthForm{}
}

func (f *SiweAuthForm) ParseAndValidate(r *http.Request) (*SiweAuthForm, response.Error) {
	var request *siweAuthRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetDeviceID(request, errors)
	f.validateAndSetDeviceName(request)
	f.validateAndSetMessage(request, errors)
	f.validateAndSetAddress(request, errors)
	f.validateAndSetSignature(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *SiweAuthForm) validateAndSetDeviceID(req *siweAuthRequest, errors map[string]response.ErrorMessage) {
	deviceID := strings.TrimSpace(req.DeviceID)
	if deviceID == "" {
		errors["device_id"] = response.MissedValueError("missed value")

		return
	}

	f.DeviceID = deviceID
}

func (f *SiweAuthForm) validateAndSetDeviceName(request *siweAuthRequest) {
	deviceName := strings.TrimSpace(request.DeviceName)
	if deviceName == "" {
		return
	}

	f.DeviceName = deviceName
}

func (f *SiweAuthForm) validateAndSetMessage(request *siweAuthRequest, errors map[string]response.ErrorMessage) {
	message := strings.TrimSpace(request.Message)
	if message == "" {
		errors["message"] = response.MissedValueError("missed value")

		return
	}

	f.Message = message
}

func (f *SiweAuthForm) validateAndSetAddress(request *siweAuthRequest, errors map[string]response.ErrorMessage) {
	address := strings.TrimSpace(request.Address)
	if address == "" {
		errors["address"] = response.MissedValueError("missed value")

		return
	}

	if !common.IsHexAddress(address) {
		errors["address"] = response.MissedValueError("invalid address")

		return
	}

	f.Address = common.HexToAddress(address)
}

func (f *SiweAuthForm) validateAndSetSignature(request *siweAuthRequest, errors map[string]response.ErrorMessage) {
	signature := strings.TrimSpace(request.Signature)
	if signature == "" {
		errors["signature"] = response.MissedValueError("missed value")

		return
	}

	f.Signature = signature
}
