package utils

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.100 Safari/537.36"

type Config struct {
	Url                  string // url to get
	Spoof                bool   // spoof user-agent
	MaxErrors            int    // max number of errors to sustain before admitting defeat
	NoBackoff            bool   // don't backoff on retry (only relevant if maxErrors > 0)
	Outfile              string // download url to file (otherwise return as string)
	DontRetryOnBadStatus bool   // if HTTP status code is not 200, don't retry
}

func Wget(conf Config) string {
	client := &http.Client{}
	startingBackoff := 100 * time.Millisecond
	if conf.NoBackoff {
		startingBackoff = 0
	}

	req, err := http.NewRequest("GET", conf.Url, nil)
	if err != nil {
		log.Fatalln(err)
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
			log.Fatalln(err)
		}
		startingBackoff *= 2 // exponential backoff
		time.Sleep(startingBackoff)
	}

	// return the response
	if conf.Outfile == "" {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		return string(body)
	}

	// output the response to file, instead of returning it
	os.MkdirAll(conf.Outfile, os.ModePerm)
	file, err := os.Create(conf.Outfile)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	file.Close()
	return ""
}
