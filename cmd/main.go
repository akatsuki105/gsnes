package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pokemium/gsnes/core"
)

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

	if *isDebug {
		go e.runServer(3001)
	}

	ebiten.SetWindowTitle("gsnes")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(e); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitCodeError
	}

	return ExitCodeOK
}

func printVersion() {
	fmt.Println(title+":", version)
}

func (e *emulator) runServer(port int) {
	fmt.Printf("Server listening on port %d\n", port)
	http.HandleFunc("/pause", cors(e.pause))
	http.HandleFunc("/resume", cors(e.resume))
	http.HandleFunc("/step", cors(e.step))
	http.HandleFunc("/wram", cors(e.wram))
	http.HandleFunc("/vram", cors(e.vram))
	http.HandleFunc("/palette", cors(e.palette))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func cors(fn func(w http.ResponseWriter, req *http.Request)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		fn(w, req)
	}
}

func (e *emulator) pause(w http.ResponseWriter, req *http.Request) {
	e.sfc.Pause(true)
}

func (e *emulator) resume(w http.ResponseWriter, req *http.Request) {
	e.sfc.Pause(false)
}

func (e *emulator) step(w http.ResponseWriter, req *http.Request) {
	e.sfc.Run()
}

func (e *emulator) wram(w http.ResponseWriter, req *http.Request) {
	pbuf, _ := e.sfc.MemoryBuffer("WRAM")
	buf := (*[core.WRAM_SIZE]uint8)(pbuf)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(buf[:])
}

func (e *emulator) vram(w http.ResponseWriter, req *http.Request) {
	pbuf, _ := e.sfc.MemoryBuffer("VRAM")
	buf := (*[core.VRAM_SIZE]uint8)(pbuf)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(buf[:])
}

func (e *emulator) palette(w http.ResponseWriter, req *http.Request) {
	pbuf, _ := e.sfc.MemoryBuffer("PALETTE")
	buf := (*[512]uint8)(pbuf)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(buf[:])
}
