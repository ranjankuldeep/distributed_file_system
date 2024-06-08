package main

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

const defaultRootFolderName = "ggnetwork"

type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	Filename string
}

func CASPathTransformFunc(key string) PathKey {
	// Here only file name is getting hashed. Not the file contents.
	hash := sha1.Sum([]byte(key)) // Old way of hashing but still can be used
	hashStr := hex.EncodeToString(hash[:])

	blocksize := 5
	sliceLen := len(hashStr) / blocksize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"), // Will join the stirng of the slice and holds the final path value
		Filename: hashStr,
	}
}

type StoreOpts struct {
	// Root is the folder name of the root, containing all the folders/files of the system.
	Root              string
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

// It represents the abstraction of hasing or fetching ivolved.
type Store struct {
	StoreOpts
}
