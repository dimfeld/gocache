package gocache

import (
	//"github.com/dimfeld/simpleblog/lru"
	"errors"
	"strings"
	"sync"
)

type MemoryCache struct {
	lock        sync.RWMutex
	memoryLimit int // Memory Limit in bytes. This isn't quite accurate.
	objectLimit int
	object      map[string]Object
	memoryUsage int
}

// Set adds the given item to the cache. Objects whose data exceeds
// the object size limit are silently not added, but an existing object
// with the same key will be deleted in this case.
//
// If adding an object to the cache would exceed the total memory limit,
// the cache will be trimmed first. Currently, the trim process evicts all
// items from the cache.
func (m *MemoryCache) Set(path string, item Object) error {
	m.lock.Lock()

	oldItem, ok := m.object[path]
	if ok {
		m.memoryUsage -= len(oldItem.Data)
	}

	if m.objectLimit != 0 && len(item.Data) > m.objectLimit {
		// If this item takes up more than our object limit, don't store it in this cache.
		delete(m.object, path)
	} else {
		if m.memoryUsage+len(item.Data) > m.memoryLimit {
			m.trim()
		}

		m.object[path] = item
		m.memoryUsage += len(item.Data)
	}
	m.lock.Unlock()

	return nil
}

func (m *MemoryCache) Del(path string) {
	if strings.HasSuffix(path, "*") {
		if path == "*" {
			// Simple case first: everything is being deleted.
			m.object = make(map[string]Object)
			m.memoryUsage = 0
			return
		}

		// Delete all matching objects in the cache.
		// This is slow, and where a radix tree might be better.
		// But it also doesn't come up too often.
		prefix := path[0 : len(path)-1]
		m.lock.Lock()
		for key, item := range m.object {
			if strings.HasPrefix(key, prefix) {
				m.memoryUsage -= len(item.Data)
				delete(m.object, key)
			}
		}
		m.lock.Unlock()
	} else {
		m.lock.Lock()
		m.memoryUsage -= len(m.object[path].Data)
		delete(m.object, path)
		m.lock.Unlock()
	}
}

func (m *MemoryCache) Get(path string, filler Filler) (item Object, err error) {
	m.lock.RLock()
	item, ok := m.object[path]
	m.lock.RUnlock()

	if !ok {
		if filler != nil {
			return filler.Fill(m, path)
		} else {
			return item, errors.New("Item not found")
		}
	}

	return item, nil
}

// Trim memory usage of the array. Right now this just clears all the data, which is obviously
// non-optimal. Once the LRU list is written, it will use that instead.
// This function assumes that we already have a write lock.
func (m *MemoryCache) trim() {
	m.object = make(map[string]Object)
	m.memoryUsage = 0
}

// NewMemoryCache creates a new cache, with the given limits for total memory and
// object size. The objectLimit may be set to 0 for no limit.
func NewMemoryCache(memoryLimit int, objectLimit int) *MemoryCache {
	return &MemoryCache{memoryLimit: memoryLimit,
		objectLimit: objectLimit,
		object:      make(map[string]Object)}
}
