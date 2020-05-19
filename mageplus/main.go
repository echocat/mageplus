package mageplus

import (
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
	if err := godotenv.Load(".env", ".env.build", ".env.mage", ".env.example"); err != nil {
		errlog := log.New(stderr, "", 0)
		errlog.Println("Error:", err)
		return 2
	}
	return mage.ParseAndRun(stdout, stderr, stdin, args)
}
