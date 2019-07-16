/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package bitshares

import (
	"sync"
	"time"

	"github.com/blocktree/openwallet/openwallet"
)

type CacheManager struct {
	bucket map[string]interface{}
	mutex  *sync.RWMutex
}

func NewCacheManager() CacheManager {
	cm := CacheManager{
		bucket: make(map[string]interface{}),
		mutex:  new(sync.RWMutex),
	}
	return cm
}

func (cm *CacheManager) Add(key string, value interface{}, duration time.Duration) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.bucket[key] = value
	return nil
}
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	value := cm.bucket[key]
	if value == nil {
		return nil, false
	}
	return value, true
}
func (cm *CacheManager) GetCacheEntry(key string) (*openwallet.CacheEntry, bool) {
	return nil, true
}
func (cm *CacheManager) Remove(key string) (interface{}, bool) {
	return nil, true
}
func (cm *CacheManager) Contains(key string) bool {
	return true
}
func (cm *CacheManager) Clear() {}
