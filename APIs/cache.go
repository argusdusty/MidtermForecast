package APIs

import (
	. "MidtermForecast/Utils"
	"bytes"
	"io"
	"net/http"
	"os"
	"time"
)

func LoadCache(url, file string, maxAge time.Duration) *bytes.Reader {
	stat, err := os.Stat(file)
	if err != nil || (maxAge > 0 && time.Now().Sub(stat.ModTime()) > maxAge) {
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		f, err := os.Create(file)
		if err != nil {
			panic(err)
		}
		io.Copy(f, resp.Body)
		f.Close()
		resp.Body.Close()
	}
	r, _ := LoadFileCache(file)
	return r
}
