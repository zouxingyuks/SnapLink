package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"testing"
)

var options = &redis.Options{
	Addr:     "100.76.246.116:6379",
	Password: "",
	DB:       0,
}

func TestNewBloomFilterCache(t *testing.T) {
	client := redis.NewClient(options)
	NewBloomFilterCache(client)
}

func Test_bloomFilterCache_Create(t *testing.T) {
	client := redis.NewClient(options)
	bf := NewBloomFilterCache(client)
	if err := bf.Create(context.Background(), "test", 0.01, 1000); err != nil {
		t.Error(err)
	}
}

func Test_bloomFilterCache_Add(t *testing.T) {
	client := redis.NewClient(options)
	bf := NewBloomFilterCache(client)
	if err := bf.Add(context.Background(), "test", "1"); err != nil {
		t.Error(err)
	}
}

func Test_bloomFilterCache_MAdd(t *testing.T) {
	client := redis.NewClient(options)
	bf := NewBloomFilterCache(client)
	if err := bf.MAdd(context.Background(), "test", "1", "2", "3"); err != nil {
		t.Error(err)
	}
}

func Test_bloomFilterCache_Exists(t *testing.T) {
	client := redis.NewClient(options)
	bf := NewBloomFilterCache(client)
	if exists, err := bf.Exists(context.Background(), "test", "1"); err != nil {
		t.Error(err)
	} else {
		t.Log(exists)
	}
}

func Test_bloomFilterCache_MExists(t *testing.T) {
	client := redis.NewClient(options)
	bf := NewBloomFilterCache(client)
	if exists, err := bf.MExists(context.Background(), "test", "1", "2", "3"); err != nil {
		t.Error(err)
	} else {
		t.Log(exists)
	}
}

func Test_bloomFilterCache_Delete(t *testing.T) {
	client := redis.NewClient(options)
	bf := NewBloomFilterCache(client)
	if err := bf.Delete(context.Background(), "test"); err != nil {
		t.Error(err)
	}
}
