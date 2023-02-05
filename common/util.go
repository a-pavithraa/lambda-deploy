package common

import (
	"fmt"
	"strings"
)

type InputError struct {
	Message string
}

func (e *InputError) Error() string {
	return fmt.Sprintf(
		"Error in inputs: %s",
		e.Message)
}

func TrimAndCheckEmptyString(s *string) bool {
	*s = strings.TrimSpace(*s)
	return len(*s) == 0
}
