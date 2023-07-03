package request

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	OffsetField   = "offset"
	DefaultOffset = 0

	LimitField   = "limit"
	DefaultLimit = 20
)

var (
	ErrInvalidArguments = errors.New("invalid request arguments")
)

func ExtractPagination(r *http.Request) (offset int, limit int, err error) {
	offset, err = extractIntFromQuery(r.URL.Query(), OffsetField, DefaultOffset)
	if err != nil {
		return DefaultOffset, DefaultLimit, fmt.Errorf("%w: %v", ErrInvalidArguments, err)
	}

	limit, err = extractIntFromQuery(r.URL.Query(), LimitField, DefaultLimit)
	if err != nil {
		return DefaultOffset, DefaultLimit, fmt.Errorf("%w: %v", ErrInvalidArguments, err)
	}

	return offset, limit, nil
}

func extractIntFromQuery(query url.Values, field string, defaultValue int) (int, error) {
	str := strings.TrimSpace(query.Get(field))
	if str == "" {
		return defaultValue, nil
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue, err
	}

	return val, nil
}
