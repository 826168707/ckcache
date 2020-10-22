package ckcache

import "sync"

var (
	cache = make(map[string]*CacheTable)
	mutex sync.RWMutex
)

// 找到 name 对应的 cacheTable,如果不存在就创建一个
func Cache(name string) *CacheTable {
	mutex.RLock()
	table,ok := cache[name]
	mutex.RUnlock()

	// 确保准确性再加写锁找一次
	if !ok {
		mutex.Lock()
		table,ok = cache[name]
		if !ok {
			table = &CacheTable{
				name: name,
				items: make(map[interface{}]*CacheItem),
			}
			cache[name] = table
		}
		mutex.Unlock()
	}

	return table
}
