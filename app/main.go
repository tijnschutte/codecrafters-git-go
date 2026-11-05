package main

import (
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}
	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		if len(os.Args) != 4 {
			fmt.Fprintf(os.Stderr, "usage: mygit cat-file -p <blob_sha>\n")
			os.Exit(1)
		}
		objectHash := os.Args[3]
		if len(objectHash) != 40 {
			fmt.Fprintf(os.Stderr, "Object's hash must have 40 characters")
			os.Exit(1)
		}
		path := ".git/objects/" + objectHash[:2] + "/" + objectHash[2:]
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "Content not found")
			os.Exit(1)
		}
		file, _ := os.Open(path)
		c, _ := zlib.NewReader(file)
		defer c.Close()
		data, _ := io.ReadAll(c)
		_, content, ok := strings.Cut(strings.TrimPrefix(string(data), "blob "), "\x00")
		if !ok {
			fmt.Fprintf(os.Stderr, "Content not found")
			os.Exit(1)
		}
		fmt.Print(content)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
