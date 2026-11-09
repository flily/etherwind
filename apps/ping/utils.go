package ping

import (
	"errors"
	"os"
)

func rootError(err error) error {
	for {
		next := errors.Unwrap(err)
		if next == nil {
			return err
		}
		err = next
	}
}

func isPermissionDenied(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, os.ErrPermission) {
		return true
	}

	return false
}
