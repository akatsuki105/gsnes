package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/akatsuki105/gsnes/core"
	"github.com/pkg/profile"
)

// ExitCode represents program's status code
type exitCode int

// exit code
const (
	ExitCodeOK exitCode = iota
	ExitCodeError
)

var (
	s = flag.Int("s", 30, "How many seconds to run the emulator.")
)

func main() {
	os.Exit(int(run()))
}

func run() exitCode {
	defer profile.Start(profile.ProfilePath("./build/profiler")).Stop()

	flag.Parse()
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

	c := core.New()
	c.LoadROM(romData)

	fmt.Printf("Run emulator for %d seconds\n", *s)
	for i := 0; i < (*s)*60; i++ {
		c.RunFrame()
		c.FrameBuffer()
		if i%60 == 0 {
			fmt.Printf("%d sec\n", i/60+1)
		}
	}

	return ExitCodeOK
}
