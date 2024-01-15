package cache

import (
	"testing"
	"time"

	"SnapLink/internal/model"

	"github.com/zhufuyi/sponge/pkg/gotest"
	"github.com/zhufuyi/sponge/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func newShortLinkCache() *gotest.Cache {
	record1 := &model.ShortLink{}
	record1.ID = 1
	record2 := &model.ShortLink{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(uint64(record1.ID)): record1,
		utils.Uint64ToStr(uint64(record2.ID)): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewShortLinkCache(&model.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_shortLinkCache_Set(t *testing.T) {
	c := newShortLinkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ShortLink)
	err := c.ICache.(ShortLinkCache).Set(c.Ctx, uint64(record.ID), record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ShortLinkCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_shortLinkCache_Get(t *testing.T) {
	c := newShortLinkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ShortLink)
	err := c.ICache.(ShortLinkCache).Set(c.Ctx, uint64(record.ID), record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ShortLinkCache).Get(c.Ctx, uint64(record.ID))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ShortLinkCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_shortLinkCache_MultiGet(t *testing.T) {
	c := newShortLinkCache()
	defer c.Close()

	var testData []*model.ShortLink
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ShortLink))
	}

	err := c.ICache.(ShortLinkCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ShortLinkCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ShortLink))
	}
}

func Test_shortLinkCache_MultiSet(t *testing.T) {
	c := newShortLinkCache()
	defer c.Close()

	var testData []*model.ShortLink
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ShortLink))
	}

	err := c.ICache.(ShortLinkCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_shortLinkCache_Del(t *testing.T) {
	c := newShortLinkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ShortLink)
	err := c.ICache.(ShortLinkCache).Del(c.Ctx, uint64(record.ID))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_shortLinkCache_SetCacheWithNotFound(t *testing.T) {
	c := newShortLinkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ShortLink)
	err := c.ICache.(ShortLinkCache).SetCacheWithNotFound(c.Ctx, uint64(record.ID))
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewShortLinkCache(t *testing.T) {
	c := NewShortLinkCache(&model.CacheType{
		CType: "memory",
	})

	assert.NotNil(t, c)
}
