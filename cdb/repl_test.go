package cdb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/moltencan/funnyfarm/cdb"
	"github.com/stretchr/testify/assert"
)

func TestRepl(t *testing.T) {
	// this really should not work
	assert := assert.New(t)
	files := []string{"db1", "db2"}

	for _, d := range files {
		_, err := os.Create(d)
		assert.Nil(err)
	}

	err := os.Mkdir("sync1", 0755)
	assert.Nil(err)

	defer func() {
		for _, d := range files {
			err := os.Remove(d)
			assert.Nil(err)
		}

		err = os.RemoveAll("sync1")
		assert.Nil(err)
	}()

	dba, err := cdb.NewDB("db1")
	assert.Nil(err)
	assert.NotNil(dba)

	dba, err = cdb.StartMaster("db1", "sync1")
	assert.Nil(err)
	assert.NotNil(dba)

	dbb, err := cdb.StartSlave("db2", "sync1")
	assert.Nil(err)
	assert.NotNil(dbb)

	testData := []byte("test data")
	id := dba.Set(testData)

	id2, ok := dbb.Get(id)
	assert.True(ok)
	assert.Equal(testData, id2)

	go func() {
		for i := 0; i < 200; i++ {
			data, ok := dbb.Get(1)
			assert.True(ok)
			assert.Equal([]byte("test data"), data)
		}
	}()

	for i := 0; i < 100; i++ {
		tdStr := fmt.Sprintf("test data %d", i)
		dba.Set([]byte(tdStr))
	}

	id2, ok = dbb.Get(100)
	assert.True(ok)
	assert.Equal([]byte("test data 98"), id2)
}
