package common

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const layout = "2006-01-02T15:04:05Z"

type Time struct {
	*time.Time
}

func NewTime(t time.Time) *Time {
	return &Time{
		&t,
	}
}

func (t *Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	if s == "null" {
		t.Time = &time.Time{}
		return nil
	}

	var parsed time.Time
	err := json.Unmarshal(b, &parsed)
	if err == nil {
		t.Time = &parsed

		return nil
	}

	parsed, err = time.Parse(layout, s)
	if err != nil {
		return err
	}

	t.Time = &parsed

	return nil
}

func (t *Time) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}

	return []byte(fmt.Sprintf(`"%s"`, t.Time.UTC().Format(layout))), nil
}

func (t *Time) AsTime() *time.Time {
	ptr := *t.Time

	return &ptr
}
