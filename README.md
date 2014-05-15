gocache [![Build Status](https://travis-ci.org/dimfeld/gocache.png?branch=master)](https://travis-ci.org/dimfeld/gocache) [![GoDoc](http://godoc.org/github.com/dimfeld/gocache?status.png)](http://godoc.org/github.com/dimfeld/gocache)
=======

Go library with various caching functionality



## Interface
The Cache interface defines Get, Set, and Del operations, allowing cache objects to be interchanged easily.

For each of these functions, the key is a string and the value is the Object type, which contains a []byte slice with the data and a timestamp.


## Functionality

### MemoryCache
This is a simple cache that stores objects in memory. It has an optional object size
limit and a total size limit, and will evict items when the total memory usage
exceeds the limit.

### DiskCache
This is a cache that stores objects on disk under a given directory. The
key passed to the cache forms the path, and any subdirectories present in the path
are created.

The DiskCache.ScanExisting function populates the cache's item list from the existing files in
the cache directory, if desired.

### MultiLevel
MultiLevel is a wrapper for multiple Caches. It starts with the first cache,
and on a miss queries the upper levels. When an upper-level cache hits, the item
is filled into the lower-level caches.

### SplitSize
SplitSize is a wrapper for multiple Caches that delegates objects of different sizes
to different caches. This can be used, for example, with two MemoryCache objects
to ensure that a few large objects will not evict the smaller objects in the cache.
