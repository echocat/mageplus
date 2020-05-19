package mageplus

import (
	"errors"
	"flag"
	"fmt"
	mio "github.com/echocat/mageplus/io"
	"github.com/echocat/mageplus/sdk"
	"github.com/echocat/mageplus/wrapper"
	"github.com/joho/godotenv"
	"github.com/magefile/mage/mage"
	"github.com/magefile/mage/mg"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
)

const (
	initFile              = "magefile.go"
	Wrapper  mage.Command = 1000
	notSet                = "<not set>"
)

var (
	// set by ldflags when you "mageplus build"
	commitHash = notSet
	timestamp  = notSet
	gitTag     = notSet

	DotEnvFilesCandidates = []string{".env.mage", ".env.build", ".env", ".env.example"}

	debug = log.New(ioutil.Discard, "DEBUG: ", log.Ltime|log.Lmicroseconds)

	initOutput = template.Must(template.New("").Parse(mageTpl))
)

type Invocation struct {
	mage.Invocation
	EnsureSdk bool // If true SDK will be ensured and on demand downloaded
}

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
	errlog := log.New(stderr, "", 0)
	out := log.New(stdout, "", 0)
	inv, cmd, err := Parse(stderr, stdout, args)
	inv.Stderr = stderr
	inv.Stdin = stdin
	if err == flag.ErrHelp {
		return 0
	}
	if err != nil {
		errlog.Println("Error:", err)
		return 2
	}

	if files, err := resolveDotEnvFiles(); err != nil {
		errlog.Println("Error:", err)
		return 2
	} else if len(files) == 0 {
		// Ignore
	} else if err := godotenv.Load(files...); err != nil {
		errlog.Println("Error:", err)
		return 2
	}

	switch cmd {
	case mage.Version:
		out.Println("Mage Build Tool", gitTag)
		out.Println("Build Date:", timestamp)
		out.Println("Commit:", commitHash)
		out.Println("built with:", runtime.Version())
		return 0
	case mage.Init:
		if err := generateInit(inv.Dir); err != nil {
			errlog.Println("Error:", err)
			return 1
		}
		out.Println(initFile, "created")
		return 0
	case Wrapper:
		version := gitTag
		//noinspection GoBoolExpressions
		if version == notSet {
			if len(inv.Args) < 1 {
				errlog.Println("Error:", "version of wrapper required as argument")
				return 1
			}
			version = inv.Args[0]
		}
		if err := wrapper.Write(inv.Dir, version); err != nil {
			errlog.Println("Error:", err)
			return 1
		}
		out.Println("mageplusw", "created")
		return 0
	case mage.Clean:
		if err := removeContents(inv.CacheDir); err != nil {
			out.Println("Error:", err)
			return 1
		}
		out.Println(inv.CacheDir, "cleaned")
		return 0
	case mage.CompileStatic:
		if err := EnsureSdkIfRequired(inv); err != nil {
			errlog.Println("Error:", err)
			return 1
		}
		return mage.Invoke(inv.Invocation)
	case mage.None:
		if err := EnsureSdkIfRequired(inv); err != nil {
			errlog.Println("Error:", err)
			return 1
		}
		return mage.Invoke(inv.Invocation)
	default:
		panic(fmt.Errorf("unknown command type: %v", cmd))
	}
}

func EnsureSdkIfRequired(inv Invocation) error {
	if !inv.EnsureSdk {
		return nil
	}
	s, err := sdk.Discover()
	if err != nil {
		return err
	}
	if err := os.Setenv("GOROOT", s.Root); err != nil {
		return err
	}
	path := os.Getenv("PATH")
	if err := os.Setenv("PATH", fmt.Sprintf("%s%c%s", filepath.Dir(s.GoBinary), os.PathListSeparator, path)); err != nil {
		return err
	}
	return nil
}

// Parse parses the given args and returns structured data.  If parse returns
// flag.ErrHelp, the calling process should exit with code 0.
func Parse(stderr, stdout io.Writer, args []string) (inv Invocation, cmd mage.Command, err error) {
	inv.Stdout = stdout
	fs := flag.FlagSet{}
	fs.SetOutput(stdout)

	// options flags

	fs.BoolVar(&inv.Force, "f", false, "force recreation of compiled magefile")
	fs.BoolVar(&inv.Debug, "debug", mg.Debug(), "turn on debug messages")
	fs.BoolVar(&inv.EnsureSdk, "ensuresdk", true, "will ensure a working golang SDK")
	fs.BoolVar(&inv.Verbose, "v", mg.Verbose(), "show verbose output when running mage targets")
	fs.BoolVar(&inv.Help, "h", false, "show this help")
	fs.DurationVar(&inv.Timeout, "t", 0, "timeout in duration parsable format (e.g. 5m30s)")
	fs.BoolVar(&inv.Keep, "keep", false, "keep intermediate mage files around after running")
	fs.StringVar(&inv.Dir, "d", ".", "run magefiles in the given directory")
	fs.StringVar(&inv.GoCmd, "gocmd", mg.GoCmd(), "use the given go binary to compile the output")
	fs.StringVar(&inv.GOOS, "goos", "", "set GOOS for binary produced with -compile")
	fs.StringVar(&inv.GOARCH, "goarch", "", "set GOARCH for binary produced with -compile")

	// commands below

	fs.BoolVar(&inv.List, "l", false, "list mage targets in this directory")
	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "show version info for the mageplus binary")
	var mageInit bool
	fs.BoolVar(&mageInit, "init", false, "create a starting template if no mage files exist")
	var ensureWrapper bool
	fs.BoolVar(&ensureWrapper, "wrapper", false, "ensures a wrapper with the version of this mageplus binary")
	var clean bool
	fs.BoolVar(&clean, "clean", false, "clean out old generated binaries from CACHE_DIR")
	var compileOutPath string
	fs.StringVar(&compileOutPath, "compile", "", "output a static binary to the given path")

	fs.Usage = func() {
		_, _ = fmt.Fprint(stdout, `
mageplus [options] [target]

MagePlus is a make-like command runner.  See https://github.com/echocat/mageplus for full docs.

Commands:
  -clean     clean out old generated binaries from CACHE_DIR
  -compile <string>
             output a static binary to the given path
  -init      create a starting template if no mage files exist
  -wrapper   ensures a wrapper with the version of this mageplus binary
  -l         list mage targets in this directory
  -h         show this help
  -version   show version info for the mageplus binary

Options:
  -d <string> 
             run magefiles in the given directory (default ".")
  -debug     turn on debug messages
  -ensuresdk will ensure a working golang SDK (default: true)
  -h         show description of a target
  -f         force recreation of compiled magefile
  -keep      keep intermediate mage files around after running
  -gocmd <string>
		     use the given go binary to compile the output (default: "go")
  -goos      sets the GOOS for the binary created by -compile (default: current OS)
  -goarch    sets the GOARCH for the binary created by -compile (default: current arch)
  -t <string>
             timeout in duration parsable format (e.g. 5m30s)
  -v         show verbose output when running mage targets
`[1:])
	}
	err = fs.Parse(args)
	if err == flag.ErrHelp {
		// parse will have already called fs.Usage()
		return inv, cmd, err
	}
	if err == nil && inv.Help && len(fs.Args()) == 0 {
		fs.Usage()
		// tell upstream, to just exit
		return inv, cmd, flag.ErrHelp
	}

	numCommands := 0
	switch {
	case mageInit:
		numCommands++
		cmd = mage.Init
	case ensureWrapper:
		numCommands++
		cmd = Wrapper
	case compileOutPath != "":
		numCommands++
		cmd = mage.CompileStatic
		inv.CompileOut = compileOutPath
		inv.Force = true
	case showVersion:
		numCommands++
		cmd = mage.Version
	case clean:
		numCommands++
		cmd = mage.Clean
		if fs.NArg() > 0 {
			// Temporary dupe of below check until we refactor the other commands to use this check
			return inv, cmd, errors.New("-h, -init, -wrapper, -clean, -compile and -version cannot be used simultaneously")

		}
	}
	if inv.Help {
		numCommands++
	}

	if inv.Debug {
		debug.SetOutput(stderr)
	}

	inv.CacheDir = mg.CacheDir()

	if numCommands > 1 {
		debug.Printf("%d commands defined", numCommands)
		return inv, cmd, errors.New("-h, -init, -wrapper, -clean, -compile and -version cannot be used simultaneously")
	}

	if cmd != mage.CompileStatic && (inv.GOARCH != "" || inv.GOOS != "") {
		return inv, cmd, errors.New("-goos and -goarch only apply when running with -compile")
	}

	inv.Args = fs.Args()
	if inv.Help && len(inv.Args) > 1 {
		return inv, cmd, errors.New("-h can only show help for a single target")
	}

	if len(inv.Args) > 0 && cmd != mage.None {
		return inv, cmd, fmt.Errorf("unexpected arguments to command: %q", inv.Args)
	}
	inv.HashFast = mg.HashFast()
	return inv, cmd, err
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

func generateInit(dir string) error {
	debug.Println("generating default magefile in", dir)
	f, err := os.Create(filepath.Join(dir, initFile))
	if err != nil {
		return fmt.Errorf("could not create mage template: %v", err)
	}
	defer mio.CloseQuietly(f)

	if err := initOutput.Execute(f, nil); err != nil {
		return fmt.Errorf("can't execute magefile template: %v", err)
	}

	return nil
}

// removeContents removes all files but not any subdirectories in the given
// directory.
func removeContents(dir string) error {
	debug.Println("removing all files in", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		err = os.Remove(filepath.Join(dir, f.Name()))
		if err != nil {
			return err
		}
	}
	return nil

}
