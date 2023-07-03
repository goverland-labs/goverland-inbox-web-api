package helpers

import (
	"encoding/json"
	"io"
)

func ReadJSON[T any](source io.ReadCloser, v *T) error {
	body, err := io.ReadAll(source)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := json.Unmarshal(body, &v); err != nil {
		return err
	}

	return nil
}
