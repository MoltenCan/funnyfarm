package cdb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/moltencan/funnyfarm/cdb"
	"github.com/stretchr/testify/assert"
)

var tFile = "./cdb_test.db"

func TestCDB(t *testing.T) {
	defer func() {
		err := os.Remove(tFile)
		assert.Nil(t, err)
	}()

	assert := assert.New(t)

	// test new db
	c, err := cdb.NewDB(tFile)
	assert.Nil(err)
	assert.NotNil(c)

	// create some data
	newID := c.Set([]byte("test"))
	assert.Equal(uint64(1), newID)

	newID2 := c.Set([]byte("test2"))
	assert.Equal(uint64(2), newID2)

	newID3 := c.SetWithKV([]byte("test3"), map[string]string{"test": "test"})
	assert.Equal(uint64(3), newID3)

	newID4 := c.SetWithKV([]byte("test3"), map[string]string{"test": "bob"})
	assert.Equal(uint64(4), newID4)

	// save
	err = c.Save()
	assert.Nil(err)

	// make sure file was created
	stat, err := os.Stat(tFile)
	if err != nil {
		fmt.Println(err)
	}
	assert.Nil(err)
	assert.NotNil(stat)

	// testing basic get
	r1, ok := c.Get(newID)
	assert.True(ok)
	assert.Equal([]byte("test"), r1)

	r2, ok := c.GetEntry(newID2)
	assert.True(ok)
	assert.Equal([]byte("test2"), r2.Val)
	assert.Equal(map[string]string{}, r2.KV)
	assert.Equal(uint64(2), r2.ID)

	r3, ok := c.GetEntry(newID3)
	assert.True(ok)
	assert.Equal([]byte("test3"), r3.Val)
	assert.Equal(map[string]string{"test": "test"}, r3.KV)
	assert.Equal(uint64(3), r3.ID)

	// make sure we can load the db back in
	c, err = cdb.LoadDB(tFile)
	assert.Nil(err)
	assert.NotNil(c)

	// testing basic get post re-load
	rx, ok := c.GetEntry(newID3)
	assert.True(ok)
	assert.Equal([]byte("test3"), rx.Val)
	assert.Equal(map[string]string{"test": "test"}, rx.KV)
	assert.Equal(uint64(3), rx.ID)

	// testing update keys
	err = c.UpdateKVs(newID, "aaa")
	assert.EqualError(err, "invalid kvPair")

	err = c.UpdateKVs(newID, "aaa", "bbb")
	assert.Nil(err)

	rx, ok = c.GetEntry(newID)
	assert.True(ok)
	assert.Equal([]byte("test"), rx.Val)
	assert.Equal(map[string]string{"aaa": "bbb"}, rx.KV)
	assert.Equal(uint64(1), rx.ID)

	err = c.UpdateKVs(newID3, "ccc", "ddd")
	assert.Nil(err)

	rx, ok = c.GetEntry(newID3)
	assert.True(ok)
	assert.Equal([]byte("test3"), rx.Val)
	assert.Equal(map[string]string{"ccc": "ddd", "test": "test"}, rx.KV)
	assert.Equal(uint64(3), rx.ID)

	// test search
	resp := c.GetWhere("nothing", "ddd")
	assert.Len(resp, 0)

	resp = c.GetWhere("ccc", "ddd")
	assert.Len(resp, 1)

	resp = c.GetWhere("test", "test")
	assert.Len(resp, 1)

}
