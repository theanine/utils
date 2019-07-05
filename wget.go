package utils

import (
	"encoding/gob"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36"

type Config struct {
	Cache                time.Duration // cache time, default to 0 (don't cache)
	Force                bool          // force download even if cached
	Url                  string        // url to get
	Spoof                bool          // spoof user-agent
	MaxErrors            int           // max number of errors to sustain before admitting defeat
	NoBackoff            bool          // don't backoff on retry (only relevant if maxErrors > 0)
	Outfile              string        // download url to file (otherwise return as string)
	DontRetryOnBadStatus bool          // if HTTP status code is not 200, don't retry
}

type cached struct {
	Expiry  time.Time
	Content string
}

var cache map[string]cached // url -> content (+expiry)

func loadCache() error {
	cache = make(map[string]cached)

	decodeFile, err := os.Open(".cached")
	if err != nil {
		return nil
	}
	defer decodeFile.Close()

	decoder := gob.NewDecoder(decodeFile)
	decoder.Decode(&cache)
	return nil
}

func saveCache(url string, body string, d time.Duration) error {
	expiry := time.Now().Add(d)
	cache[url] = cached{expiry, body}

	encodeFile, err := os.Create(".cached")
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(encodeFile)
	if err := encoder.Encode(cache); err != nil {
		return err
	}
	encodeFile.Close()
	return nil
}

func remoteGet(conf Config) (string, error) {
	client := &http.Client{}
	startingBackoff := 100 * time.Millisecond
	if conf.NoBackoff {
		startingBackoff = 0
	}

	req, err := http.NewRequest("GET", conf.Url, nil)
	if err != nil {
		return "", err
	}

	if conf.Spoof {
		req.Header.Set("User-Agent", userAgent)
	}

	var resp *http.Response
	var errs int
	for {
		resp, err = client.Do(req)
		if err == nil && (resp.StatusCode == 200 || conf.DontRetryOnBadStatus) {
			// we got the response, and it's either a 200 or we don't care what the status code is
			defer resp.Body.Close()
			break
		}
		errs++
		if errs > conf.MaxErrors {
			return "", err
		}
		startingBackoff *= 2 // exponential backoff
		time.Sleep(startingBackoff)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	body := string(bytes)

	// save response to cache
	if conf.Cache > 0 {
		saveCache(conf.Url, body, conf.Cache)
	}
	return body, nil
}

func cacheGet(conf Config) string {
	var cached string
	var expiry time.Time
	if v, ok := cache[conf.Url]; ok {
		cached = v.Content
		expiry = v.Expiry
	}
	if cached == "" || time.Until(expiry) < 0 {
		return ""
	}
	return cached
}

func Wget(conf Config) (string, error) {
	if cache == nil {
		loadCache()
	}

	var body string
	if !conf.Force {
		body = cacheGet(conf)
	}

	var err error
	if body == "" {
		if body, err = remoteGet(conf); err != nil {
			return "", err
		}
	}

	// return the response
	if conf.Outfile == "" {
		return body, nil
	}

	// output the response to file, instead of returning it
	dir, _ := filepath.Split(conf.Outfile)
	os.MkdirAll(dir, os.ModePerm)
	file, err := os.Create(conf.Outfile)
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(body)
	file.Close()
	return "", err
}
