package Utils

import (
	"io"
	"os"
	"sync"
	"time"
)

type RawObject struct {
	Raw    []byte
	Object interface{}
}

type FileCacheObject struct {
	Data interface{} // Raw bytes of file
	Time time.Time
}

var (
	FileCache = new(sync.Map)
	fclock    sync.RWMutex
)

// Load a file from the FileCache. If the file is not in the cache, save it there.
// No memory limit: Be aware of your memory requirements when using this.
func LoadFileCache(file string, parser func(io.Reader) interface{}) (data interface{}, t time.Time) {
	f := func() {
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
			stat, err := f.Stat()
			if err != nil {
				panic(err)
			}
			data = parser(f)
			t = stat.ModTime()
			FileCache.Store(file, FileCacheObject{Data: data, Time: t})
			return
		} else {
			fc := fc_raw.(FileCacheObject)
			data = fc.Data
			t = fc.Time
			return
		}
	}
	defer func() {
		if r := recover(); r != nil {
			time.Sleep(5 * time.Second) // May have been in the process of writing. Give it one more try
			f()
		}
	}()
	f()
	return
}
