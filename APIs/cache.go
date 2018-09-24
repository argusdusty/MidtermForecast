package APIs

import (
	. "MidtermForecast/Utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func LoadCache(url, file string, maxAge time.Duration, parser func(r io.Reader) interface{}) interface{} {
	stat, err := os.Stat(file)
	if err != nil || (maxAge > 0 && time.Now().Sub(stat.ModTime()) > maxAge) {
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		os.MkdirAll(filepath.Dir(file), 0755)
		f, err := os.Create(file)
		if err != nil {
			panic(err)
		}
		io.Copy(f, resp.Body)
		f.Close()
		resp.Body.Close()
	}
	r, _ := LoadFileCache(file, parser)
	return r
}
