package main

import (
	"ckcache"
	"fmt"
	"strconv"
)

func main() {
	cache := ckcache.Cache("myCache")

	cache.SetValueLoader(func(key interface{}, args ...interface{}) *ckcache.CacheItem {
		val := "This is a test with key" + key.(string)
		item := ckcache.NewCacheItem(key,val,0)
		return item
		})
	for i := 0; i < 10; i++ {
		res,err := cache.Value("someKey_"+strconv.Itoa(i))
		if err == nil {
			fmt.Println("Found value in cache",res.Value())
		}else {
			fmt.Println("Error retrieving value from cache",err)
		}
	}
}


