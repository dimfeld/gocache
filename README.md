gocache
=======

Go library with various caching functionality

# Cache Functionality

## MemoryCache
This is a simple cache that stores objects in memory. It has an optional object size
limit and a total size limit, and will evict items when the total memory usage
exceeds the limit.

## DiskCache
This is a cache that stores objects on disk, below the specified directory. The
key passed to the cache forms the path, and any subdirectories present in the path
are created.

## MultiLevelCache
MultiLevelCache is a wrapper for multiple caches. It starts with the first cache,
and on a miss queries the upper levels. When an upper-level cache hits, the item
is filled into the lower-level caches.

## SplitSize
SplitSize is a wrapper for multiple caches that delegates objects of different sizes
to different caches. This can be used, for example, with multiple MemoryCache objects
to ensure that a few large objects will not evict the smaller objects in the cache.

