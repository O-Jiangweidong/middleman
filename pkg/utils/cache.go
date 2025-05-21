package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var (
	instance     *CacheManager
	instanceOnce sync.Once
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type CacheManager struct {
	db *badger.DB
	mu sync.RWMutex
}

func NewCacheManager(dataDir string) (*CacheManager, error) {
	instanceOnce.Do(func() {
		opts := badger.DefaultOptions(dataDir).
			WithInMemory(false).
			WithLogger(nil)

		db, err := badger.Open(opts)
		if err != nil {
			log.Fatalf("Cache init failed: %v\n", err)
		}
		instance = &CacheManager{db: db}
	})
	return instance, nil
}

func (c *CacheManager) Set(key string, value interface{}, expireSecond int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := int64(0)
	if expireSecond > 0 {
		expiration = time.Now().Unix() + expireSecond
	}

	item := CacheItem{Value: value, Expiration: expiration}

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func (c *CacheManager) Get(key string, dest interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var item *badger.Item
	err := c.db.View(func(txn *badger.Txn) error {
		var err error
		item, err = txn.Get([]byte(key))
		return err
	})
	// key 不存在
	if err != nil {
		return err
	}

	var cacheItem CacheItem
	err = item.Value(func(val []byte) error {
		return json.Unmarshal(val, &cacheItem)
	})
	if err != nil {
		return err
	}

	if cacheItem.Expiration > 0 && time.Now().Unix() > cacheItem.Expiration {
		c.mu.RUnlock()
		c.mu.Lock()
		defer c.mu.Unlock()

		err = c.db.View(func(txn *badger.Txn) error {
			item, err = txn.Get([]byte(key))
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				if err := json.Unmarshal(val, &cacheItem); err != nil {
					return err
				}
				if cacheItem.Expiration > 0 && time.Now().Unix() > cacheItem.Expiration {
					return txn.Delete([]byte(key))
				}
				return nil
			})
		})
		if err == nil {
			return errors.New("缓存已过期")
		}
		if errors.Is(err, badger.ErrKeyNotFound) {
			return errors.New("缓存已过期")
		}
		return err
	}

	return json.Unmarshal([]byte(fmt.Sprint(cacheItem.Value)), dest)
}

func (c *CacheManager) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (c *CacheManager) Clear() error {
	return c.db.DropAll()
}

func (c *CacheManager) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Close()
}

func GetCache() *CacheManager {
	cache, err := NewCacheManager("data")
	if err != nil {
		log.Fatal("Init cache component failed.")
	}
	return cache
}
