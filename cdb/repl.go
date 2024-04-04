package cdb

import (
	"bytes"
	"encoding/gob"

	"github.com/moltencan/funnyfarm/fsio"
)

// a fool may try to use this as some kind of replacted db

// start db as master or slave
// if master push the db to fsio and pray

func StartMaster(dbloc, syncloc string) (*charDB, error) {
	db, err := LoadDB(dbloc)
	if err != nil {
		return nil, err
	}

	// the local db must always save on the master
	db.saveHook = func(s *Store) error {
		if err := db.saveLocal(); err != nil {
			return err
		}

		buffer := bytes.NewBuffer([]byte{})
		enc := gob.NewEncoder(buffer)
		err = enc.Encode(s)
		if err != nil {
			return err
		}
		return fsio.Push(syncloc, buffer.Bytes())
	}

	// and now kick off a save so the slave has something to load and we know we can save to the syncloc
	if err := db.Save(); err != nil {
		return nil, err
	}

	return db, nil
}

func StartSlave(dbloc, syncloc string) (*charDB, error) {
	db := newDB(dbloc)

	go func() {
		fsio.Pull(syncloc, func(data []byte) error {
			db.mtx.Lock()
			defer db.mtx.Unlock()

			dec := gob.NewDecoder(bytes.NewBuffer(data))
			err := dec.Decode(&db.store)
			if err != nil {
				return err
			}

			return db.save()
		})
	}()

	return db, nil
}
