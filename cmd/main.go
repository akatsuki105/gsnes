package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pokemium/gsnes/core"
)

var exits = []func(){}

var version string

const (
	title = "gsnes"
)

// ExitCode represents program's status code
type exitCode int

// exit code
const (
	ExitCodeOK exitCode = iota
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
	os.Exit(int(run()))
}

// Run program
func run() exitCode {
	var (
		showVersion = flag.Bool("v", false, "show version")
		showRomInfo = flag.Bool("r", false, "show rom info")
		isDebug     = flag.Bool("d", false, "debug mode")
	)

	flag.Parse()
	if *showVersion {
		printVersion()
		return ExitCodeOK
	}

	// main routine
	romPath := flag.Arg(0)
	if romPath == "" {
		fmt.Fprintln(os.Stderr, "rom path is required")
		return ExitCodeError
	}

	romData, err := os.ReadFile(romPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitCodeError
	}

	if *showRomInfo {
		core.PrintCartInfo(romData)
		return ExitCodeOK
	}

	e := new()
	e.setDebugMode(*isDebug)
	e.sfc.LoadROM(romData)

	ebiten.SetWindowTitle("gsnes")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	defer func() {
		for _, f := range exits {
			f()
		}
	}()

	if err := ebiten.RunGame(e); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitCodeError
	}

	return ExitCodeOK
}

func printVersion() {
	fmt.Println(title+":", version)
}
