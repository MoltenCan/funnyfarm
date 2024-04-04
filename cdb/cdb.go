package cdb

import (
	"encoding/gob"
	"fmt"
	"os"
	"sync"
)

type DBEntry struct {
	ID  uint64
	Val []byte
	KV  map[string]string
}

type Store struct {
	DB     map[uint64]DBEntry
	LastID uint64
}

type charDB struct {
	FileLocation string
	AutoSave     bool

	store    Store
	mtx      sync.RWMutex
	saveHook func(*Store) error
}

func newDB(location string) *charDB {
	return &charDB{
		FileLocation: location,
		store: Store{
			DB:     make(map[uint64]DBEntry),
			LastID: 0,
		},
		AutoSave: true,
	}
}

func NewDB(location string) (*charDB, error) {
	c := newDB(location)
	return c, c.Save()
}

func LoadDB(location string) (*charDB, error) {
	r, err := os.Open(location)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	c := newDB(location)

	dec := gob.NewDecoder(r)
	err = dec.Decode(&c.store)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (x *charDB) Save() error {
	x.mtx.Lock()
	defer x.mtx.Unlock()
	return x.save()
}

func (x *charDB) save() error {
	defer BCCount(BCStart())

	// run save hook
	if x.saveHook != nil {
		return x.saveHook(&x.store)
	}
	return x.saveLocal()
}

func (x *charDB) saveLocal() error {
	// open file
	fh, err := os.OpenFile(x.FileLocation, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fh.Close()

	// gob it out
	enc := gob.NewEncoder(fh)
	err = enc.Encode(x.store)
	if err != nil {
		return err
	}

	return nil
}

func (x *charDB) nextID() uint64 {
	x.store.LastID++
	return x.store.LastID
}

func (x *charDB) Set(v []byte) uint64 {
	x.mtx.Lock()
	defer x.mtx.Unlock()
	return x.set(v)
}

func (x *charDB) set(v []byte) uint64 {
	id := x.nextID()
	x.store.DB[id] = DBEntry{
		ID:  id,
		Val: v,
		KV:  make(map[string]string),
	}
	x.autoSave()
	return id
}

func (x *charDB) Update(id uint64, v []byte) uint64 {
	x.mtx.Lock()
	defer x.mtx.Unlock()

	if e, ok := x.store.DB[id]; ok {
		e.Val = v
		x.store.DB[id] = e
		x.autoSave()
		return id
	}
	return x.set(v)
}
func (x *charDB) UpdateKVs(id uint64, kvPair ...string) error {
	if len(kvPair)%2 != 0 {
		return fmt.Errorf("invalid kvPair")
	}
	x.mtx.Lock()
	defer x.mtx.Unlock()

	if e, ok := x.store.DB[id]; ok {
		for i := 0; i < len(kvPair); i += 2 {
			e.KV[kvPair[i]] = kvPair[i+1]
		}
		x.store.DB[id] = e
		x.autoSave()
		return nil
	}
	return fmt.Errorf("id not found")
}

func (x *charDB) SetWithKV(v []byte, kv map[string]string) uint64 {
	x.mtx.Lock()
	defer x.mtx.Unlock()

	id := x.nextID()
	x.store.DB[id] = DBEntry{
		ID:  id,
		Val: v,
		KV:  kv,
	}
	x.autoSave()
	return id
}

func (x *charDB) Delete(id uint64) {
	x.mtx.Lock()
	defer x.mtx.Unlock()
	delete(x.store.DB, id)
	x.autoSave()
}

func (x *charDB) autoSave() {
	if x.AutoSave {
		x.save()
	}
}

func (x *charDB) Get(id uint64) ([]byte, bool) {
	x.mtx.RLock()
	defer x.mtx.RUnlock()

	if _, ok := x.store.DB[id]; !ok {
		return nil, false
	}

	return x.store.DB[id].Val, true
}

// GetEntry returns a copy of the DBEntry for the given id
func (x *charDB) GetEntry(id uint64) (DBEntry, bool) {
	x.mtx.RLock()
	defer x.mtx.RUnlock()

	if _, ok := x.store.DB[id]; !ok {
		return DBEntry{}, false
	}

	return x.store.DB[id], true
}

func (x *charDB) GetWhere(key, val string) [][]byte {
	x.mtx.RLock()
	defer x.mtx.RUnlock()

	var res [][]byte
	for _, v := range x.store.DB {
		if v.KV[key] == val {
			res = append(res, v.Val)
		}
	}
	return res
}
