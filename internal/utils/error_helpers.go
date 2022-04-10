package utils

import "fmt"

func FormatError(err error, msg string) error {
	return fmt.Errorf("[%w] %s", err, msg)
}
