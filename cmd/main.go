package main

import (
	"flag"
	"fmt"
	"os"
)

var version string

const (
	title = "gsnes"
)

// ExitCode represents program's status code
type ExitCode int

// exit code
const (
	ExitCodeOK ExitCode = iota
	ExitCodeError
)

func init() {
	if version == "" {
		version = "develop"
	}

	flag.Usage = func() {
		usage := fmt.Sprintf(`Usage:
    %s [arg] [input]
input: a filepath
Arguments: 
`, title)

		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
}

func main() {
	os.Exit(int(Run()))
}

// Run program
func Run() ExitCode {
	var (
		showVersion = flag.Bool("v", false, "show version")
	)

	flag.Parse()
	if *showVersion {
		printVersion()
		return ExitCodeOK
	}

	// main routine
	return ExitCodeOK
}

func printVersion() {
	fmt.Println(title+":", version)
}
