# `profiler`

Take profile data for emulation core using `profile` package.

## Usage

```sh
make build-profiler
./build/profiler/profiler -s=30 ROM_PATH
go tool pprof -http=":8081" ./build/profiler/profiler ./build/profiler/cpu.pprof
# or: go tool pprof -png ./build/profiler/profiler ./build/profiler/cpu.pprof > ./build/profiler/cpu.pprof.png
```
