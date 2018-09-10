package Utils

import (
	"bytes"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type FileCacheObject struct {
	Data []byte // Raw bytes of file
	Time time.Time
}

var (
	FileCache = new(sync.Map)
	fclock    sync.RWMutex
)

// Load a file from the FileCache. If the file is not in the cache, save it there.
// No memory limit: Be aware of your memory requirements when using this.
func LoadFileCache(file string) (*bytes.Reader, time.Time) {
	fclock.RLock()
	fc_raw, ok := FileCache.Load(file)
	fclock.RUnlock()
	var load bool
	if !ok {
		load = true
	} else {
		fc := fc_raw.(FileCacheObject)
		stat, err := os.Stat(file)
		if err != nil || (stat.ModTime().After(fc.Time)) {
			load = true
		}
	}
	if load {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
		stat, err := f.Stat()
		if err != nil {
			panic(err)
		}
		FileCache.Store(file, FileCacheObject{Data: data, Time: stat.ModTime()})
		return bytes.NewReader(data), stat.ModTime()
	} else {
		fc := fc_raw.(FileCacheObject)
		return bytes.NewReader(fc.Data), fc.Time
	}
}
