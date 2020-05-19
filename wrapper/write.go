package wrapper

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
)

var (
	unixScript    = ``
	windowsScript = ``
)

func Write(targetDir string, version string) error {
	//noinspection GoBoolExpressions
	if unixScript == "" || windowsScript == "" {
		panic("unixScript and/or windowsScript are still empty. resources.go not generated before building?")
	}
	unixScriptFile := filepath.Join(targetDir, "mageplusw")
	windowsScriptFile := filepath.Join(targetDir, "mageplusw.cmd")
	if unixScriptFileExists, err := exists(unixScriptFile); err != nil {
		return err
	} else if err := writeFile(unixScriptFile, unixScript, version, 0755); err != nil {
		return err
	} else if err := writeFile(windowsScriptFile, windowsScript, version, 0644); err != nil {
		return err
	} else {
		if unixScriptFileExists {
			noticeAfterCreation(unixScriptFile)
		}
		return nil
	}
}

func writeFile(target string, rawBase64EncodedContent string, version string, perm os.FileMode) error {
	if content, err := prepareContent(rawBase64EncodedContent, version); err != nil {
		return err
	} else if err := createDirectorsForFileIfRequired(target); err != nil {
		return err
	} else if f, err := openFile(target, perm); err != nil {
		return err
	} else {
		defer f.Close()
		_, err := f.Write(content)
		return err
	}
}

func prepareContent(rawBase64EncodedContent string, version string) ([]byte, error) {
	if b, err := base64.RawURLEncoding.DecodeString(rawBase64EncodedContent); err != nil {
		return nil, err
	} else {
		return []byte(strings.Replace(string(b), "####VERSION####", version, -1)), nil
	}
}

func createDirectorsForFileIfRequired(file string) error {
	return os.MkdirAll(filepath.Dir(file), 0755)
}

func openFile(file string, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, perm)
}

func exists(file string) (bool, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}
