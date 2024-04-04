package fsio

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
)

// a really simple and crappy way to sync data between two processes using a directory
// please use a ramfs or something similar for the directory otherwise it's gonna shred your disk
// also, this is not a good way to do this, but it's a simple way to do this

const (
	SuffixData = ".data"
	SuffixDone = ".done"
)

type PullFunc func([]byte) error

func Pull(dir string, f PullFunc) error {

	fsn, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer fsn.Close()

	if err := fsn.Add(dir); err != nil {
		return err
	}

	for ev := range fsn.Events {
		// only respond to create events
		if ev.Op&fsnotify.Create != fsnotify.Create {
			continue
		}
		if filepath.Ext(ev.Name) == SuffixDone {
			fPath := ev.Name[:len(ev.Name)-len(SuffixDone)] + SuffixData
			data, err := os.ReadFile(fPath)
			if err != nil {
				return fmt.Errorf("read file error: %w", err)
			}
			if err := f(data); err != nil {
				return err
			}
			os.Remove(ev.Name)
			os.Remove(fPath)
		}
	}
	return nil
}

func Push(dir string, data []byte) error {
	f := strconv.FormatInt(time.Now().UnixMicro(), 10)

	// write the data
	fPath := filepath.Join(dir, f+SuffixData)
	fh, err := os.OpenFile(fPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err = fh.Write(data); err != nil {
		return err
	}
	fh.Sync()
	fh.Close()

	// write the done
	fPath = filepath.Join(dir, f+SuffixDone)
	fh, err = os.OpenFile(fPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	fh.Sync()
	fh.Close()

	return nil
}
