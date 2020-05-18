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

func Exists(filename string) (bool, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("cannot get information if file '%s': %v", filename, err)
	} else {
		return true, nil
	}
}

func DirExists(filename string) (bool, error) {
	if fi, err := os.Stat(filename); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("cannot get information if file '%s': %v", filename, err)
	} else {
		return fi.IsDir(), nil
	}
}

func FileExists(filename string) (bool, error) {
	if fi, err := os.Stat(filename); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("cannot get information if file '%s': %v", filename, err)
	} else {
		return !fi.IsDir(), nil
	}
}
