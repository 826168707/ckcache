package ckcache

import (
	"errors"
	"sync"
	"time"
)



type CacheTable struct {
	sync.RWMutex

	// 表名
	name		string

	// 所有缓存 item
	items 		map[interface{}]*CacheItem

	// 用来触发清除的 Timer
	cleanupTimer	*time.Timer

	// 触发清除的周期
	cleanupInterval	time.Duration

	// 当试图读取一个不存在的 key 时触发的回调方法
	loadValue func(key interface{},args ...interface{}) *CacheItem

	// 添加 item 时触发的回调方法
	addedItem []func(item *CacheItem)

	// 删除 item 是触发的回调方法
	aboutToDeleteItem []func(item *CacheItem)

}

// 计算表中 item 数目
func (table *CacheTable) Count() int {
	table.RLock()
	defer table.RUnlock()
	return len(table.items)
}

// 遍历所有 Key　item 到自定义方法中
func (table *CacheTable) Foreach(trans func(key interface{}, item *CacheItem)) {
	table.RLock()
	defer table.RUnlock()
	for k, item := range table.items {
		trans(k,item)
	}
}

//  过期检查	通过在 table 中维护一个 Timer ，
//	每次触发时循环读取所有 item 信息并删除过期的 item，然后修改 cleanupInterval 为最近一个 item 距离过期的时间，
//	然后到期时创建一个 goroutine 去触发一个新的过期检查，本次检查结束
func (table *CacheTable) expirationCheck() {
	table.Lock()
	defer table.Unlock()

	if table.cleanupTimer != nil  {		// 因为其他方法也可能会主动调用 expirationCheck ，因此要主动停止 Timer
		table.cleanupTimer.Stop()
	}

	now := time.Now()
	smallestDuration := 0 * time.Second
	for key, item := range table.items {
		item.RLock()
		lifespan := item.lifespan
		accessedOn := item.accessedOn
		item.RUnlock()

		if lifespan == 0 {
			continue
		}

		if now.Sub(accessedOn) >= lifespan {		// 过期了
			table.deleteInternal(key)
		}else {										// 没过期，比较距离过期时间
			if smallestDuration == 0 || lifespan - now.Sub(accessedOn) < smallestDuration {
				smallestDuration = lifespan - now.Sub(accessedOn)
			}
		}
	}

	// 设置 Timer 下一次的触发时间
	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck()
		})
	}
}

// 添加 item
func (table *CacheTable) addInternal(item *CacheItem) {
	table.items[item.key] = item
	expDur := table.cleanupInterval
	addedItem := table.addedItem
	table.Unlock()

	// 触发 addedInternal
	if addedItem != nil {
		for _,callback := range addedItem {
			callback(item)
		}
	}

	// 检查是否要刷新 expirationCheck
	if item.lifespan > 0 && (expDur == 0 || item.lifespan < expDur) {
		table.expirationCheck()
	}
}

func (table *CacheTable) Add(key interface{}, lifespan time.Duration, value interface{}) *CacheItem {
	item := NewCacheItem(key,value,lifespan)
	table.Lock()
	table.addInternal(item)

	return item
}


// 删除 item
func (table *CacheTable) deleteInternal(key interface{}) (*CacheItem, error) {
	item,ok := table.items[key]
	if !ok {
		return nil,errors.New("Key is not found")
	}
	aboutDeleteItem := table.aboutToDeleteItem
	table.Unlock()

	// 触发 aboutDeleteItem
	if aboutDeleteItem != nil {
		for _,callback := range aboutDeleteItem {
			callback(item)
		}
	}

	// 触发 aboutToExpire
	item.RLock()
	if item.aboutToExpire != nil {
		for _, callback := range item.aboutToExpire {
			callback(key)
		}
	}
	item.RUnlock()

	table.Lock()
	delete(table.items,key)

	return item,nil
}

func (table *CacheTable) Delete(key interface{}) (*CacheItem, error) {
	table.Lock()
	defer table.Unlock()
	return table.deleteInternal(key)
}



// 查询 item 是否存在
func (table *CacheTable) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()

	_,ok := table.items[key]
	return ok
}

// 查询 item 是否存在，不存在就添加   存在返回 false 不存在返回 true
func (table *CacheTable) NotFoundAdd(key interface{}, lifespan time.Duration, value interface{}) bool {
	table.Lock()
	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}
	item := NewCacheItem(key,value,lifespan)
	table.addInternal(item)
	return true
}


// 根据 Key 获取对应 item
func (table *CacheTable) Value(key interface{},args ...interface{}) (*CacheItem, error) {
	table.RLock()
	item,ok := table.items[key]
	loadValue := table.loadValue
	table.RUnlock()

	if ok {
		item.KeepAlive()
		return item,nil
	}

	// 找不到，尝试重新添加
	if loadValue != nil {
		item := loadValue(key,args...)
		if item != nil {
			table.Add(key,item.lifespan,item.value)
			return item,nil
		}
		return nil,errors.New("loadValue is null")
	}

	return nil,errors.New("item is not found")
}



// 格式化 cacheTable
func (table *CacheTable) Flush() {
	table.Lock()
	table.items = make(map[interface{}]*CacheItem)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	table.Unlock()
}


// 设置 loadValue 回调方法，该方法会在试图获取一个不存在的 item 时触发
func (table *CacheTable) SetValueLoader(f func(interface{}, ...interface{}) *CacheItem)  {
	table.Lock()
	defer table.Unlock()
	table.loadValue = f
}

func (table *CacheTable) RemoveAddedItemCallback() {
	table.Lock()
	table.addedItem = nil
	table.Unlock()
}


// 设置 addItem 回调方法，该方法会在添加一个新的 item 时触发
func (table *CacheTable) SetAddedItemCallback(f func(item *CacheItem)) {
	if len(table.addedItem) > 0 {
		table.RemoveAddedItemCallback()
	}
	table.Lock()
	table.addedItem = append(table.addedItem,f)
	table.Unlock()
}

func (table *CacheTable) AddAddedItemCallback(f func(item *CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.addedItem = append(table.addedItem,f)
}




// aboutToDeleteItem 回调方法相关

func (table *CacheTable) RemoveAboutToDeleteItemCallback() {
	table.Lock()
	table.aboutToDeleteItem = nil
	table.Unlock()
}

func (table *CacheTable) SetAboutToDeleteItemCallback(f func(item *CacheItem)) {
	if len(table.aboutToDeleteItem) > 0 {
		table.RemoveAddedItemCallback()
	}

	table.Lock()
	defer table.Unlock()
	table.aboutToDeleteItem = append(table.aboutToDeleteItem,f)
}

func (table *CacheTable) AddAboutToDeleteItemCallback(f func(item *CacheItem)) {
	table.Lock()
	table.aboutToDeleteItem = append(table.aboutToDeleteItem,f)
	table.Unlock()
}