package pushes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/helpers"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

type CustomPushRequest struct {
	Title         string `json:"title"`
	Body          string `json:"body"`
	ImageURL      string `json:"image_url"`
	CustomPayload string `json:"custom_payload"`
}

type CustomPushForm struct {
	Title         string
	Body          string
	ImageURL      string
	CustomPayload []byte
}

func NewCustomPushForm() *CustomPushForm {
	return &CustomPushForm{}
}

func (f *CustomPushForm) ParseAndValidate(r *http.Request) (*CustomPushForm, response.Error) {
	var request *CustomPushRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetTitle(request, errors)
	f.validateAndSetBody(request, errors)
	f.validateAndSetImageURL(request, errors)
	f.validateAndSetPayload(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *CustomPushForm) validateAndSetTitle(req *CustomPushRequest, errors map[string]response.ErrorMessage) {
	title := req.Title
	if title == "" {
		title = "default title"
	}

	f.Title = title
}

func (f *CustomPushForm) validateAndSetBody(req *CustomPushRequest, errors map[string]response.ErrorMessage) {
	body := req.Body
	if body == "" {
		body = "default body"
	}

	f.Body = body
}

func (f *CustomPushForm) validateAndSetImageURL(req *CustomPushRequest, errors map[string]response.ErrorMessage) {
	imageURL := req.ImageURL
	if imageURL == "" {
		imageURL = "https://dummyimage.com/64x64/ad88ad/212422.jpg&text=gov"
	}

	f.ImageURL = imageURL
}

func (f *CustomPushForm) validateAndSetPayload(req *CustomPushRequest, errors map[string]response.ErrorMessage) {
	var pl []byte
	if req.CustomPayload != "" {
		pl = []byte(req.CustomPayload)
	} else {
		data := map[string]string{
			"custom":   "data",
			"app":      "goverland",
			"added_at": time.Now().String(),
		}
		pl, _ = json.Marshal(data)
	}

	f.CustomPayload = pl
}
