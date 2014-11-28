/**
 * Ebase frame for daemon program
 * Author Jonsen Yang
 * Date 2013-07-05
 * Copyright (c) 2013 ForEase Times Technology Co., Ltd. All rights reserved.
 */
package ebase

import (
    "sync"
)

type Cache struct {
    Item map[interface{}]interface{}
    Lock *sync.RWMutex
    Len uint64
}

// 新建一个cache
func NewCache() (c *Cache) {
    c = &Cache{
        Item: make(map[interface{}]interface{}),
        Lock: new(sync.RWMutex),
    }

    return
}

// 从cache里获取
func (c *Cache)Get(key interface{}) interface{} {
    c.Lock.RLock()
    defer c.Lock.RUnlock()

    if val, ok := c.Item[key]; ok {
        return val
    }

    return nil
}

// 检查cache是否存在
func (c *Cache)Exists( key interface{}) bool {
    _, ok := c.Item[key]
    return ok
}


// 设置cache
func (c *Cache)Set(key, val interface{})  bool {
    c.Lock.Lock()
    defer c.Lock.Unlock()

    if v, ok := c.Item[key]; ok && val == v {
        return false
    }

    c.Item[key] = val
    c.Len++

    return true
}

// 删除cache
func (c *Cache)Del(key interface{}) {
    c.Lock.Lock()
    defer c.Lock.Unlock()

    delete(c.Item, key)
    c.Len--
}

// 清除cache
// 参数是一个函数
// 传递外部函数判断cache里数据，决定是否要删除
func (c *Cache)Cleanup(f func(interface{}) bool) {
    for key, value := range c.Item {
        if f(value) {
            c.Del(key)
        }
    }
}
