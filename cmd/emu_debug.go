package main

import (
	"fmt"
	"log"
	"os"

	"github.com/edsrzf/mmap-go"
)

func (e *emulator) MMap(path string, size int) []uint8 {
	f := createEmptyFile(path, size)
	exits = append(exits, func() {
		if err := f.Close(); err == nil {
			fmt.Println("Close ", path)
		}
	})

	m, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		panic(err)
	}
	return m
}

func createEmptyFile(path string, size int) *os.File {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]uint8, size)
	f.Seek(0, 0)
	f.Write(buf)
	return f
}
