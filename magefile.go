//+build mage

// This is the build script for Mage. The install target is all you really need.
// The release target is for generating official releases and is really only
// useful to project admins.
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	errLog     = log.New(os.Stderr, "", 0)
	releaseTag = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	version    = func() (result string) {
		if s, err := sh.Output("git", "describe", "--tags"); err == nil {
			result = s
		}
		if s, ok := os.LookupEnv("TAG"); ok {
			result = s
		}
		if result == "" {
			return "dev"
		}
		if !releaseTag.MatchString(result) {
			errLog.Fatalf("TAG environment variable must be in semver vx.x.x format, but was %s", result)
		}
		return
	}()
)

func GenerateSources() error {
	mg.Deps(GenerateWrapperSources)
	return nil
}

// Runs "go install" for mage.  This generates the version info the binary.
func Install() error {
	mg.Deps(GenerateSources)

	name := "mageplus"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	gocmd := mg.GoCmd()
	// use GOBIN if set in the environment, otherwise fall back to first path
	// in GOPATH environment string
	bin, err := sh.Output(gocmd, "env", "GOBIN")
	if err != nil {
		return fmt.Errorf("can't determine GOBIN: %v", err)
	}
	if bin == "" {
		gopath, err := sh.Output(gocmd, "env", "GOPATH")
		if err != nil {
			return fmt.Errorf("can't determine GOPATH: %v", err)
		}
		paths := strings.Split(gopath, string([]rune{os.PathListSeparator}))
		bin = filepath.Join(paths[0], "bin")
	}
	// specifically don't mkdirall, if you have an invalid gopath in the first
	// place, that's not on us to fix.
	if err := os.Mkdir(bin, 0700); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create %q: %v", bin, err)
	}
	path := filepath.Join(bin, name)

	// we use go build here because if someone built with go get, then `go
	// install` turns into a no-op, and `go install -a` fails on people's
	// machines that have go installed in a non-writeable directory (such as
	// normal OS installs in /usr/bin)
	return sh.RunV(gocmd, "build", "-o", path, "-ldflags="+flags(), "github.com/echocat/mageplus")
}

// Generates a new release.  Expects the TAG environment variable to be set,
// which will create a new tag with that name.
func Release() (err error) {
	if _, ok := os.LookupEnv("TAG"); !ok {
		return errors.New("TAG envar required")
	}
	mg.Deps(GenerateSources)

	if err := downloadDependency(); err != nil {
		return err
	}

	if err := sh.RunV("git", "tag", "-a", version, "-m", version); err != nil {
		return err
	}
	if err := sh.RunV("git", "push", "origin", version); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			sh.RunV("git", "tag", "--delete", version)
			sh.RunV("git", "push", "--delete", "origin", version)
		}
	}()
	return sh.RunV("goreleaser", "--rm-dist")
}

func downloadDependency() error {
	oldWd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir("/"); err != nil {
		return err
	}
	defer os.Chdir(oldWd)
	if err := sh.RunV("go", "get", "-u", "github.com/goreleaser/goreleaser"); err != nil {
		return err
	}
	return nil
}

// Remove the temporarily generated files from Release.
func Clean() error {
	return sh.Rm("dist")
}

func flags() string {
	timestamp := time.Now().Format(time.RFC3339)
	hash := hash()
	return fmt.Sprintf(`-X "github.com/echocat/mageplus/mageplus.timestamp=%s" -X "github.com/echocat/mageplus/mageplus.commitHash=%s" -X "github.com/echocat/mageplus/mageplus.gitTag=%s"`, timestamp, hash, version)
}

// hash returns the git hash for the current repo or "" if none.
func hash() string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return hash
}

func GenerateWrapperSources() error {
	wf, err := ioutil.ReadFile("wrapper/mageplusw")
	if err != nil {
		return err
	}

	wcf, err := ioutil.ReadFile("wrapper/mageplusw.cmd")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile("wrapper/resources.go", []byte(fmt.Sprintf("package wrapper\n\n"+
		"func init() {\n"+
		"\tunixScript = `%s`\n"+
		"\twindowsScript = `%s`\n"+
		"}\n", string(wf), string(wcf),
	)), 0644); err != nil {
		return err
	}

	return nil
}
