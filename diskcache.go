package gocache

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
)

// DiskCache stores objects on a filesystem under a given directory.
// The key passed to the cache forms the path, and any subdirectories present in the path are created.
type DiskCache struct {
	lock     sync.RWMutex
	baseDir  string
	fileList map[string]int
}

func (d *DiskCache) Set(pathStr string, object Object) error {
	pathStr = path.Join(d.baseDir, pathStr)
	dirPath, _ := path.Split(pathStr)

	err := os.MkdirAll(dirPath, 0700)
	if err != nil {
		return fmt.Errorf("Could not create directory %s: %s", dirPath, err.Error())
	}

	err = ioutil.WriteFile(pathStr, object.Data, 0600)
	if err != nil {
		return fmt.Errorf("Could not write file %s: %s", pathStr, err.Error())
	}

	os.Chtimes(pathStr, object.ModTime, object.ModTime)
	d.lock.Lock()
	d.fileList[pathStr] = len(object.Data)
	d.lock.Unlock()

	return nil
}

// Del removes the given item from the cache, using a shell glob to
// perform the file matching.
func (d *DiskCache) Del(pathStr string) {
	pathStr = path.Join(d.baseDir, pathStr)
	matches, err := filepath.Glob(pathStr)
	if err != nil {
		return
	}

	// When recursively deleting directories from the cache, it's possible
	// that d.fileList will become out of sync with reality. But that's ok
	// since it just exists to prevent disk access on cache misses, so will
	// just result in a bit of extra disk access.

	d.lock.Lock()
	if pathStr == "*" {
		// Due to the above comment, special case when wiping out the entire
		// cache.
		d.fileList = make(map[string]int)
	}

	for _, match := range matches {
		delete(d.fileList, match)
		err = os.RemoveAll(match)
		if err != nil {
			// log removal failure
		}
	}
	d.lock.Unlock()

}

func (d *DiskCache) Get(filename string, filler Filler) (Object, error) {
	cachePath := path.Join(d.baseDir, filename)
	d.lock.RLock()
	_, ok := d.fileList[cachePath]
	d.lock.RUnlock()
	if !ok {
		// The object is not currently present in the disk cache. Try to generate it.
		if filler != nil {
			return filler.Fill(d, filename)
		} else {
			return Object{}, errors.New(filename)
		}
	}

	f, err := os.Open(cachePath)
	if err != nil {
		// The object should be present, but is not. Try to generate it.
		if filler != nil {
			return filler.Fill(d, filename)
		} else {
			return Object{}, errors.New(filename)
		}
	}

	defer f.Close()

	fstat, err := f.Stat()
	if err != nil {
		return Object{}, err
	}
	modTime := fstat.ModTime()

	buf := bytes.Buffer{}
	buf.Grow(int(fstat.Size()))
	_, err = buf.ReadFrom(f)
	obj := Object{buf.Bytes(), modTime}

	return obj, err
}

func (d *DiskCache) initialScanWalkFunc(filename string, info os.FileInfo, err error) error {
	d.fileList[filename] = int(info.Size())
	return nil
}

// ScanExisting finds all files under the supplied directory, and adds them to the cache's
// file list.
func (d *DiskCache) ScanExisting() {
	filepath.Walk(d.baseDir, d.initialScanWalkFunc)
}

// NewDiskCache returns a DiskCache, initialized to store its data under the
// given directory.
func NewDiskCache(baseDir string) (*DiskCache, error) {
	err := os.MkdirAll(baseDir, 0700)
	if err != nil {
		return nil, err
	}
	return &DiskCache{baseDir: baseDir, fileList: make(map[string]int)}, nil
}
