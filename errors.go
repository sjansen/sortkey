package sortkey

import "errors"

type ConfigError struct {
	message string
}

func (err *ConfigError) Error() string {
	return err.message
}

type InvalidValueError struct {
	message string
}

func (err *InvalidValueError) Error() string {
	return err.message
}

var ErrOverflow = errors.New("unable to generate larger value")
var ErrUnderflow = errors.New("unable to generate smaller value")
