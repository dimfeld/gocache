package gocache

import (
	"bytes"
	"compress/gzip"
	"strings"
	"time"
)

type Object struct {
	Data    []byte
	ModTime time.Time
}

type Filler interface {
	// Fill adds an object to the cache and also returns the object.
	// When the file may be compressed or not, Fill may add both versions to the cache, but should return
	// the requested version of the item.
	// Also, the implementation should be tolerant of a nil value for cache. In practice this does
	// not happen, but it is convenient for the tests.
	Fill(cache Cache, path string) (Object, error)
}

type Cache interface {
	// Get an object from the cache. On a miss, the function calls filler.Fill if it is non-nil.
	Get(path string, filler Filler) (Object, error)
	// Set adds an object to the cache.
	Set(path string, object Object) error
	// Delete an item from the cache. Include a "*" wildcard at the end to purge multiple items.
	Del(path string)
}

// CompressAndSet adds the given object to the cache and also adds a gzipped version of the data,
// appending ".gz" to the key for the compressed data.
func CompressAndSet(cache Cache, path string, data []byte, modTime time.Time) (uncompressed Object, compressed Object, err error) {

	compressedPath := path
	uncompressedPath := path

	if strings.HasSuffix(path, ".gz") {
		uncompressedPath = path[0 : len(path)-3]
	} else {
		compressedPath = path + ".gz"
	}

	gzBuf := new(bytes.Buffer)
	compressor, err := gzip.NewWriterLevel(gzBuf, gzip.BestCompression)
	if err != nil {
		return Object{}, Object{}, err
	}

	_, err = compressor.Write(data)
	compressor.Close()
	if err != nil {
		return Object{}, Object{}, err
	}

	compressedItem := Object{gzBuf.Bytes(), modTime}

	// Add the compressed version to the cache.
	err = cache.Set(compressedPath, compressedItem)
	if err != nil {
		return Object{}, Object{}, err
	}

	// Also add the uncompressed version.
	uncompressedItem := Object{data, modTime}
	err = cache.Set(uncompressedPath, uncompressedItem)

	return uncompressedItem, compressedItem, err
}
