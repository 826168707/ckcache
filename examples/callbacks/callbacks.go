package main

import (
	"ckcache"
	"fmt"
	"time"
)

func main() {
	cache := ckcache.Cache("myCache")

	// This callback will be triggered every time a new item
	// gets added to the cache.
	cache.SetAddedItemCallback(func(entry *ckcache.CacheItem) {
		fmt.Println("Added Callback 1:", entry.Key(), entry.Value(), entry.CreatedOn())
	})
	cache.AddAddedItemCallback(func(entry *ckcache.CacheItem) {
		fmt.Println("Added Callback 2:", entry.Key(), entry.Value(), entry.CreatedOn())
	})
	// This callback will be triggered every time an item
	// is about to be removed from the cache.
	cache.SetAboutToDeleteItemCallback(func(entry *ckcache.CacheItem) {
		fmt.Println("Deleting:", entry.Key(), entry.Value(), entry.CreatedOn())
	})

	// Caching a new item will execute the AddedItem callback.
	cache.Add("someKey", 0, "This is a test!")

	// Let's retrieve the item from the cache
	res, err := cache.Value("someKey")
	if err == nil {
		fmt.Println("Found value in cache:", res.Value())
	} else {
		fmt.Println("Error retrieving value from cache:", err)
	}

	// Deleting the item will execute the AboutToDeleteItem callback.
	cache.Delete("someKey")

	cache.RemoveAddedItemCallback()
	// Caching a new item that expires in 3 seconds
	res = cache.Add("anotherKey", 3*time.Second, "This is another test")

	// This callback will be triggered when the item is about to expire
	res.SetAboutToExpireCallback(func(key interface{}) {
		fmt.Println("About to expire:", key.(string))
	})

	time.Sleep(5 * time.Second)
}
