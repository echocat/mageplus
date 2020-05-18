package io

import (
	"fmt"
	"os"
)

func Touch(filename string, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("cannot touch '%s': %v", filename, err)
	}
	return f.Close()
}
