package mageplus

import (
	mio "github.com/echocat/mageplus/io"
	"github.com/joho/godotenv"
	"github.com/magefile/mage/mage"
	"io"
	"log"
	"os"
)

// Main is the entrypoint for running mage.  It exists external to mage's main
// function to allow it to be used from other programs, specifically so you can
// go run a simple file that run's mage's Main.
func Main() int {
	return ParseAndRun(os.Stdout, os.Stderr, os.Stdin, os.Args[1:])
}

// ParseAndRun parses the command line, and then compiles and runs the mage
// files in the given directory with the given args (do not include the command
// name in the args).
func ParseAndRun(stdout, stderr io.Writer, stdin io.Reader, args []string) int {
	errLog := log.New(stderr, "", 0)
	if files, err := resolveDotEnvFiles(); err != nil {
		errLog.Println("Error:", err)
		return 2
	} else if err := godotenv.Load(files...); err != nil {
		errLog.Println("Error:", err)
		return 2
	}
	return mage.ParseAndRun(stdout, stderr, stdin, args)
}

func resolveDotEnvFiles() ([]string, error) {
	var result []string
	for _, candidate := range DotEnvFilesCandidates {
		if exist, err := mio.FileExists(candidate); err != nil {
			return nil, err
		} else if exist {
			result = append(result, candidate)
		}
	}
	return result, nil
}

var DotEnvFilesCandidates = []string{".env", ".env.build", ".env.mage", ".env.example"}
