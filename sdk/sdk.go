package sdk

import (
	"errors"
	"github.com/blang/semver"
)

var (
	ErrSdkDifferent = errors.New("sdk different")
)

type Sdk struct {
	Version  semver.Version
	Os       string
	Arch     string
	Root     string
	GoBinary string
}

func (instance Sdk) Validate() error {
	other, err := EvalGoBinary(instance.GoBinary)
	if err == ErrNoGoBinary {
		return ErrSdkDifferent
	} else if err != nil {
		return err
	} else if !instance.Version.Equals(other.Version) {
		return ErrSdkDifferent
	} else if instance.Os != other.Os {
		return ErrSdkDifferent
	} else if instance.Arch != other.Arch {
		return ErrSdkDifferent
	}
	return nil
}
