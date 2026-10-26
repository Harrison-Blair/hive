package memory

import "errors"

type userMessaged interface {
	UserMessage() string
}

type logCategorized interface {
	LogCategory() string
}

// UserErrorText returns the exact user text of err.
func UserErrorText(err error) string {
	var messaged userMessaged
	if errors.As(err, &messaged) {
		return messaged.UserMessage()
	}
	return err.Error()
}

// LogErrorText returns the safe log text of err.
func LogErrorText(err error) string {
	var categorized logCategorized
	if errors.As(err, &categorized) {
		return categorized.LogCategory()
	}
	return DisplayValue(err.Error())
}
