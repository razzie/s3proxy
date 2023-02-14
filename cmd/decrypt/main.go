package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/razzie/s3proxy/pkg/s3proxy"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Usage: decrypt [encryption key] [filepath globs...]")
		os.Exit(0)
	}

	var filenames []string
	for _, glob := range args[1:] {
		matches, err := filepath.Glob(glob)
		if err != nil {
			fmt.Println(err)
		}
		filenames = append(filenames, matches...)
	}

	lt := s3proxy.NewLookupTable(args[0])
	count := 0
	for _, filename := range filenames {
		fi, err := os.Stat(filename)
		if err != nil {
			fmt.Println(err)
			continue
		} else if fi.IsDir() {
			continue
		}
		if decryptFile(filename, lt) {
			count++
		}
	}
	fmt.Println(count, "files encrypted")
}

func decryptFile(filename string, lt *s3proxy.LookupTable) bool {
	fmt.Printf("Decrypting file %q ... ", filename)
	file, err := os.OpenFile(filename, os.O_RDWR, 0755)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	defer file.Close()

	buffer := make([]byte, 4096)
	pos := int64(0)
	for {
		n, err := file.ReadAt(buffer, pos)
		if err != nil && err != io.EOF {
			fmt.Println("error:", err)
			return false
		}
		if n == 0 {
			break
		}

		processedChunk := buffer[:n]
		lt.Decrypt(processedChunk)
		if _, err := file.WriteAt(processedChunk, pos); err != nil {
			fmt.Println("error:", err)
			return false
		}

		pos += int64(n)
	}
	fmt.Println("done")
	return true
}
