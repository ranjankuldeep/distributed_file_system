package store

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

// Implements the path transform func type.
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
