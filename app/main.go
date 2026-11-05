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
		hash := os.Args[3]
		content, err := catFile(hash)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(content)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func catFile(objectHash string) (string, error) {
	if len(objectHash) != 40 {
		return "", errors.New("object's hash must have 40 characters")
	}
	path := ".git/objects/" + objectHash[:2] + "/" + objectHash[2:]
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("file not found: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	c, err := zlib.NewReader(file)
	if err != nil {
		return "", err
	}
	defer c.Close()

	data, err := io.ReadAll(c)
	if err != nil {
		return "", err
	}

	_, content, ok := strings.Cut(strings.TrimPrefix(string(data), "blob "), "\x00")
	if !ok {
		return "", errors.New("could not read file content")
	}

	return content, nil
}
