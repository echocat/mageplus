package sdk

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"github.com/blang/semver"
	"github.com/echocat/mageplus/http"
	mio "github.com/echocat/mageplus/io"
	"github.com/mholt/archiver/v3"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

const DefaultVersion = "1.14"
const EnvVersion = "GO_VERSION"

var errLog = log.New(os.Stderr, "", 0)
var infoLog = log.New(os.Stderr, "", 0)

type DownloadDiscovery struct {
	Version semver.Version
	Os      string
	Arch    string
}

func NewDownloadDiscovery(version string) (*DownloadDiscovery, error) {
	parsedVersion, err := semver.ParseTolerant(version)
	if err != nil {
		return nil, nil
	}
	return &DownloadDiscovery{
		Version: parsedVersion,
		Os:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}, nil
}

func MustNewDownloadDiscovery(version string) *DownloadDiscovery {
	instance, err := NewDownloadDiscovery(version)
	if err != nil {
		errLog.Fatalln("Error:", err)
	}
	return instance
}

func NewDefaultDownloadDiscovery() *DownloadDiscovery {
	if v, ok := os.LookupEnv(EnvVersion); ok {
		return MustNewDownloadDiscovery(v)
	}
	return MustNewDownloadDiscovery(DefaultVersion)
}

func (instance DownloadDiscovery) Discover() ([]Sdk, error) {
	candidate, err := instance.ToSdk()
	if err != nil {
		return nil, err
	}

	if err := candidate.Validate(); err == nil {
		return []Sdk{candidate}, nil
	} else if err == ErrSdkDifferent {
		// Continue to download it...
	} else {
		return nil, err
	}

	downloadUrl, err := instance.DownloadUrl()
	if err != nil {
		return nil, err
	}
	infoLog.Printf("Downloading Golang SDK from %s...", downloadUrl)

	if err := http.Execute(downloadUrl,
		http.WriteToTemporaryFile("", instance.String(), func(input *os.File) error {
			return instance.extract(input.Name(), candidate)
		}),
	); err != nil {
		return nil, err
	}

	return []Sdk{candidate}, nil
}

func (instance DownloadDiscovery) extract(input string, to Sdk) error {
	return instance.walker().Walk(input, func(candidate archiver.File) error {
		return instance.extractFile(candidate, to)
	})
}

func (instance DownloadDiscovery) extractFile(candidate archiver.File, to Sdk) error {
	var name string
	if h, ok := candidate.Header.(zip.FileHeader); ok {
		name = h.Name
	} else if h, ok := candidate.Header.(*zip.FileHeader); ok {
		name = h.Name
	} else if h, ok := candidate.Header.(tar.Header); ok {
		name = h.Name
	} else if h, ok := candidate.Header.(*tar.Header); ok {
		name = h.Name
	} else {
		return fmt.Errorf("unexpected header type: %v", reflect.TypeOf(candidate.Header))
	}
	if strings.HasPrefix(name, "go/") {
		name = name[3:]
	}
	if name == "" {
		return nil
	}
	target := filepath.Join(to.Root, name)
	if candidate.IsDir() {
		if err := os.MkdirAll(target, candidate.Mode()); err != nil {
			return fmt.Errorf("cannot extract '%s': %v", name, err)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("cannot create parent of '%s': %v", name, err)
	}

	w, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, candidate.Mode())
	if err != nil {
		return fmt.Errorf("cannot extract '%s': %v", name, err)
	}
	defer mio.CloseQuietly(w)

	_, err = io.Copy(w, candidate)
	if err != nil {
		return fmt.Errorf("cannot extract '%s': %v", name, err)
	}

	return nil
}

func (instance DownloadDiscovery) walker() archiver.Walker {
	if instance.Os == "windows" {
		return archiver.NewZip()
	}
	return archiver.NewTarGz()
}

func (instance DownloadDiscovery) ToSdk() (Sdk, error) {
	targetPath, err := instance.TargetPath()
	if err != nil {
		return Sdk{}, err
	}
	ext := ""
	if instance.Os == "windows" {
		ext = ".exe"
	}
	return Sdk{
		Version:  instance.Version,
		Os:       instance.Os,
		Arch:     instance.Arch,
		Root:     targetPath,
		GoBinary: filepath.Join(targetPath, "bin", "go"+ext),
	}, nil
}

func (instance DownloadDiscovery) TargetPath() (string, error) {
	gopath, err := instance.Gopath()
	if err != nil {
		return "", fmt.Errorf("cannot determine sdk target directory: %v", err)
	}
	return filepath.Join(gopath, "pkg", "sdk", instance.String()), nil
}

func (instance DownloadDiscovery) String() string {
	return fmt.Sprintf("%s.%s-%s",
		instance.VersionString(),
		instance.Os,
		instance.Arch,
	)
}

func (instance DownloadDiscovery) Gopath() (string, error) {
	if v, ok := os.LookupEnv("GOPATH"); ok {
		return v, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".go"), nil
}

func (instance DownloadDiscovery) DownloadUrl() (string, error) {
	ext := "tar.gz"
	if instance.Os == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("https://dl.google.com/go/go%s.%s",
		instance.String(),
		ext,
	), nil
}

func (instance DownloadDiscovery) VersionString() string {
	if instance.Version.Patch == 0 {
		return fmt.Sprintf("%d.%d", instance.Version.Major, instance.Version.Minor)
	}
	return fmt.Sprintf("%d.%d.%d", instance.Version.Major, instance.Version.Minor, instance.Version.Patch)
}
