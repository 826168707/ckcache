package main

import (
	"ckcache"
	"fmt"
	"time"
)

type myStruct struct {
	text 		string
	moreData 	[]byte
}


func main() {

	cache := ckcache.Cache("myCache")
	val := myStruct{
		"This is a test!",
		[]byte{},
	}
	cache.Add("someKey",5*time.Second,&val)

	res, err := cache.Value("someKey")
	if err == nil {
		fmt.Println("Found value in cache",res.Value().(*myStruct).text)
	}else {
		fmt.Println("Error retrieving value from cache",err)
	}

	time.Sleep(6 * time.Second)
	res, err = cache.Value("someKey")
	if err != nil {
		fmt.Println("Item is not cached anymore!")
	}

	cache.Add("someKey",0,&val)
	cache.SetAboutToDeleteItemCallback(func(e *ckcache.CacheItem) {
		fmt.Println("Deleting",e.Key(),e.Value().(*myStruct).text,e.CreatedOn())
	})

	cache.Delete("someKey")

	cache.Flush()
}
