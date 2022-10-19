package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	key1 := "key1"
	value1 := "value1"

	appCache := New("cacheFile.json")
	appCache.Set(key1, value1, NoExpiration)

	v, exists := appCache.Get(key1)
	if !exists {
		t.Errorf("set value not found %v", v)
	}

	if v == nil {
		t.Errorf("Value is nil, should be set")
	}

	if v != value1 {
		t.Errorf("Value should be %v, actual: %v", value1, v)
	}

	err := appCache.Delete()
	if err != nil {
		t.Errorf("Error deleting Cache: %v", err)
	}
}

func TestDate(t *testing.T) {
	key1 := "key3"
	value1 := time.Now()

	appCache := New("cacheFile2.json")

	err := appCache.Delete()
	if err != nil {
		t.Error("could not delete cache", err)
	}

	value, exists := appCache.Get("badkey")
	if exists {
		t.Errorf("Value not in cache found: %v", value)
	}

	appCache.Set(key1, value1, NoExpiration)

	v, exists := appCache.Get(key1)
	if !exists {
		t.Errorf("set value not found %v", v)
	}

	if v == nil {
		t.Errorf("Value is nil, should be set")
	}

	if v != value1 {
		t.Errorf("Value should be %v, actual: %v", value1, v)
	}

	err = appCache.Delete()
	if err != nil {
		t.Errorf("Error deleting Cache: %v", err)
	}
}

func TestExpiration(t *testing.T) {

	// key expired 1 second ago
	key1 := "key3"
	value1 := time.Now()
	expire1 := 10 * time.Millisecond

	// key expires in 2 days
	key2 := "key4"
	value2 := fmt.Sprintf("%v", time.Now())
	expire2 := 1 * time.Hour

	key3 := "key5"
	value3 := 1984
	expire3 := 1 * time.Hour

	cachePath := "cacheFile2.json"
	appCache := New(cachePath)

	appCache.Set(key1, value1, expire1)
	appCache.Set(key2, value2, expire2)
	appCache.Set(key3, value3, expire3)

	// make sure key1 is expired
	time.Sleep(15 * time.Millisecond)

	_, exists := appCache.Get(key1)
	if exists {
		t.Errorf("key %v value should not be found", key1)
	}

	value, exists := appCache.Get(key2)
	if !exists {
		t.Errorf("key %v value should be found", key2)
	}
	if value != value2 {
		t.Errorf("key %v value: %v should equal %v", key2, value, value2)
	}

	value, exists = appCache.Get(key3)
	if !exists {
		t.Errorf("key %v value should be found", key3)
	}
	if value != value3 {
		t.Errorf("key %v value: %v should equal %v", key3, value, value3)
	}

	// Recreate the cache from disk
	appCache = New(cachePath)

	_, exists = appCache.Get(key1)
	if exists {
		t.Errorf("key %v value should not be found", key1)
	}

	value, exists = appCache.Get(key2)
	if !exists {
		t.Errorf("key %v value should be found", key2)
	}
	if value != value2 {
		t.Errorf("key %v value: %v should equal %v", key2, value, value2)
	}

	value, exists = appCache.Get(key3)
	if !exists {
		t.Errorf("key %v value should be found", key3)
	}
	if value3 != int(value.(float64)) {
		t.Errorf("key %v value: %v should equal %v", key3, value, value3)
	}

	err := appCache.Delete()
	if err != nil {
		t.Errorf("Error deleting Cache: %v", err)
	}
}
