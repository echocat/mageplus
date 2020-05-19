package sdk

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/blang/semver"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	goVersionExtractPattern = regexp.MustCompile(`^go version go([0-9a-z._\-+]+) ([a-z0-9]+)\/([a-z0-9]+)$`)

	ErrNoGoBinary = errors.New("no go binary")
	ErrNoGoSdk    = errors.New("no go SDK")
)

func EvalFrom(goRoot string) (Sdk, error) {
	candidate, err := exec.LookPath(filepath.Join(goRoot, "bin", "go"))
	if errors.Unwrap(err) == exec.ErrNotFound {
		return Sdk{}, nil
	} else if err != nil {
		return Sdk{}, fmt.Errorf("cannot lookup go in '%s': %v", goRoot, err)
	}

	sdk, err := EvalGoBinary(candidate)
	if err == ErrNoGoBinary {
		return Sdk{}, ErrNoGoSdk
	} else if err != nil {
		return Sdk{}, fmt.Errorf("cannot lookup go in '%s': %v", goRoot, err)
	}

	return sdk, nil
}

func EvalFromGoroot() (Sdk, error) {
	v, ok := os.LookupEnv("GOROOT")
	if !ok {
		return Sdk{}, ErrNoGoSdk
	}
	return EvalFrom(v)
}

func EvalFromPath() (Sdk, error) {
	candidate, err := exec.LookPath("go")
	if errors.Unwrap(err) == exec.ErrNotFound {
		return Sdk{}, nil
	} else if err != nil {
		return Sdk{}, fmt.Errorf("cannot lookup go in PATH: %v", err)
	}

	sdk, err := EvalGoBinary(candidate)
	if err == ErrNoGoBinary {
		return Sdk{}, ErrNoGoSdk
	} else if err != nil {
		return Sdk{}, fmt.Errorf("cannot lookup go in PATH: %v", err)
	}

	return sdk, nil
}

func EvalGoBinary(path string) (Sdk, error) {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return Sdk{}, ErrNoGoBinary
	} else if err != nil {
		return Sdk{}, fmt.Errorf("cannot validate go binary '%s': %v", path, err)
	} else if fi.IsDir() {
		return Sdk{}, fmt.Errorf("cannot validate go binary '%s': is directory", path)
	}

	versionBuf := new(bytes.Buffer)
	c := exec.Command(path, "version")
	c.Stdout = versionBuf
	if err := c.Run(); err != nil {
		return Sdk{}, fmt.Errorf("cannot validate go binary '%s': %v", path, err)
	}

	match := goVersionExtractPattern.FindStringSubmatch(strings.TrimSpace(versionBuf.String()))
	if match == nil || len(match) != 4 {
		return Sdk{}, ErrNoGoBinary
	}

	version, err := semver.ParseTolerant(match[1])
	if err != nil {
		return Sdk{}, fmt.Errorf("cannot evaluate verison of go binary '%s': %v", path, err)
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return Sdk{}, fmt.Errorf("cannot resolve location of go binary '%s': %v", path, err)
	}
	binDir := filepath.Dir(resolved)
	if filepath.Base(binDir) != "bin" {
		return Sdk{}, fmt.Errorf("cannot resolve bin directory of go binary '%s': %v", path, err)
	}
	return Sdk{
		Version:  version,
		Os:       match[2],
		Arch:     match[3],
		Root:     filepath.Dir(binDir),
		GoBinary: path,
	}, nil
}
