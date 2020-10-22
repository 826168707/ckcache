package ckcache

import (
	"sync"
	"time"
)

type CacheItem struct {
	sync.RWMutex
	key			interface{}
	value 		interface{}
	lifespan	time.Duration	// 生命周期
	createdOn	time.Time		// 创建时间
	accessedOn	time.Time		// 最后访问时间
	accessCount	int				// 被访问的次数

	// 删除前触发的回调函数
	aboutToExpire	[]func(key interface{})
}


func NewCacheItem(key interface{}, value interface{}, lifespan time.Duration) *CacheItem {
	return &CacheItem{
		key: key,
		value: value,
		lifespan: lifespan,
		createdOn: time.Now(),
		accessedOn: time.Now(),
		accessCount: 0,
	}
}

// 刷新生命周期,通过修改 accessedOn 来刷新生命周期 （过期判定：time.Now - accessedOn > lifespan）
func (item *CacheItem) KeepAlive() {
	item.Lock()			// 修改内容要加写锁
	defer item.Unlock()
	item.accessedOn = time.Now()
	item.accessCount++
}

// 以下均为获取属性值方法

func (item *CacheItem) Lifespan() time.Duration {
	return item.lifespan
}

func (item *CacheItem) AccessedOn() time.Time {
	item.RLock()			// 读取不变的内容不用加锁，读取可变的内容要加读锁
	defer item.RUnlock()
	return item.accessedOn
}

func (item *CacheItem) CreatedOn() time.Time {
	return item.createdOn
}

func (item *CacheItem) AccessedCount() int {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

func (item *CacheItem) Key() interface{} {
	return item.key
}

func (item *CacheItem) Value() interface{} {
	return item.value
}

// 删除 aboutToExpire 的内容
func (item *CacheItem) RemoveAboutToExpireCallback() {
	item.Lock()
	item.aboutToExpire = nil
	item.Unlock()
}


// 设置删除前触发的回调函数
func (item *CacheItem) SetAboutToExpireCallback(f func(interface{})) {
	// 删除原有内容
	if len(item.aboutToExpire) > 0 {
		item.RemoveAboutToExpireCallback()
	}
	item.Lock()
	item.aboutToExpire = append(item.aboutToExpire,f)
	item.Unlock()
}

// 增加 aboutToExpire 中的回调方法
func (item *CacheItem) AddAboutToExpireCallback(f func(interface{})) {
	item.Lock()
	item.aboutToExpire = append(item.aboutToExpire,f)
	item.Unlock()
}


