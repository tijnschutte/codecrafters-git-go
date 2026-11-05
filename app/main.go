package main

import (
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		content, err := catFile(objectHash)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(content)
	case "hash-object":
		file := os.Args[3]
		data, _ := os.ReadFile(file)
		obj := fmt.Appendf([]byte("blob "), "%d\x00%s", len(data), data)
		sum := sha1.Sum(obj)
		sha := hex.EncodeToString(sum[:])

		path := fmt.Sprintf(".git/objects/%s/%s", sha[:2], sha[2:])
		os.MkdirAll(filepath.Dir(path), 0o755)
		f, _ := os.Create(path)
		defer f.Close()
		c := zlib.NewWriter(f)
		defer c.Close()
		c.Write(obj)
		fmt.Println(sha)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func catFile(sha string) (string, error) {
	if len(sha) != 40 {
		return "", errors.New("object's hash must have 40 characters")
	}
	path := fmt.Sprintf(".git/objects/%s/%s", sha[:2], sha[2:])
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
