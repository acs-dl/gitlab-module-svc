package helpers

import (
	"time"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

var formats = []string{time.RFC3339, "2006-01-02"}

func ParseTime(input string) (time.Time, error) {
	if input == "" {
		return time.Time{}, nil
	}

	for _, format := range formats {
		t, err := time.Parse(format, input)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.Errorf("Unexpected time format `%s`", input)
}
