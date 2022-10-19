package cache

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"
)

// Credit to uncomputation:
// https://uncomputation.medium.com/lets-make-a-cache-in-go-ea49fc9ebe93
// Credit to patrickmn:
// https://github.com/patrickmn/go-cache/blob/master/cache.go

const (
	// Never expire
	NoExpiration time.Duration = -1
)

type Item struct {
	Value  interface{} `json:"value"`
	Expire time.Time   `json:"expire"`
}

type Cache struct {
	path string
	data map[string]Item
}

// path to file, will load if exists
func New(path string) *Cache {

	cache := &Cache{
		path: path,
	}
	return cache
}

func (c *Cache) Get(key string) (interface{}, bool) {
	if c.data == nil {
		c.load()
	}
	item, exists := c.data[key]
	if !exists {
		return nil, false
	}

	now := time.Now()
	if item.Expire.Before(now) || item.Expire.Equal(now) {
		delete(c.data, key)
		return nil, false
	}
	return item.Value, true
}
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	if c.data == nil {
		c.load()
	}
	item := Item{value, time.Now().Add(duration)}
	c.data[key] = item
	c.save()
}

// Delete cache and remove file
func (c *Cache) Delete() error {
	_, err := os.Stat(c.path)
	if !errors.Is(err, os.ErrNotExist) {
		err = os.Remove(c.path)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (c *Cache) load() error {
	var err error
	var file *os.File

	// check if file exists
	_, err = os.Stat(c.path)
	if errors.Is(err, os.ErrNotExist) {
		c.data = make(map[string]Item)
		return nil
	}

	// file needs to be opened
	if file == nil {
		file, err = os.Open(c.path)
		if err != nil {
			return err
		}
	}

	json.NewDecoder(file).Decode(&c.data)
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) save() error {
	file, err := os.Create(c.path)
	if err != nil {
		return err
	}

	dataBytes, err := json.MarshalIndent(c.data, "", "\t")
	if err != nil {
		return err
	}

	_, err = io.Copy(file, bytes.NewReader(dataBytes))
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}
