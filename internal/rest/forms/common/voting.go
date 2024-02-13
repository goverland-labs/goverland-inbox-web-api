package common

import (
	"encoding/json"

	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type Voter string

func (v *Voter) ValidateAndSet(voter string, errors map[string]response.ErrorMessage) {
	if voter == "" {
		errors["voter"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missing voter",
		}
	}

	*v = Voter(voter)
}

type Choice json.RawMessage

func (c *Choice) ValidateAndSet(choice json.RawMessage, errors map[string]response.ErrorMessage) {
	if len(choice) == 0 {
		errors["choice"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missing choice",
		}
	}

	var skip any
	err := json.Unmarshal(choice, &skip)
	if err != nil {
		errors["choice"] = response.ErrorMessage{
			Code:    response.WrongFormat,
			Message: "choice has wrong format",
		}
	}

	*c = Choice(choice)
}
