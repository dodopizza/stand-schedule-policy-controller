package util

import (
	"errors"
)

func IgnoreMatchedError(err error, match error) error {
	if errors.Is(err, match) {
		return nil
	}
	return err
}
