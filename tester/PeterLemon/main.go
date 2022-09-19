package main

import (
	"flag"
	"fmt"
	_ "image/png"
	"os"
	"strings"
	"sync"
)

const (
	BASE_URL = "https://raw.githubusercontent.com/PeterLemon/SNES/5c5c730e754114af14faff2655cb2bb098d6419b"
	SECONDS  = 5
)

var failCount = 0

func main() {
	if err := Run(); err != nil {
		panic(err)
	}

	if failCount > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}

func Run() error {
	flag.Parse()
	testname := "ALL"
	if flag.NArg() > 0 {
		testname = flag.Arg(0)
	}
	testname = strings.ToUpper(testname)

	// テスト名は前方一致
	specified := []*testcase{}
	for i := range tests {
		ok := testname == "ALL" || strings.HasPrefix(tests[i].name, strings.ToUpper(testname))
		if ok {
			specified = append(specified, &tests[i])
		}
	}
	count := len(specified)

	// テストは並列に
	var w sync.WaitGroup
	w.Add(count)
	progress := 0
	fmt.Printf("Running tests... %d/%d", progress, count)
	for i := range specified {
		go func(i int) {
			t := specified[i]
			t.err = t.fn()
			progress++
			fmt.Printf("\rRunning tests... %d/%d", progress, count)
			w.Done()
		}(i)
	}
	w.Wait()

	// 結果
	fmt.Println()
	for i := range specified {
		fmt.Println(specified[i])
	}

	return nil
}
