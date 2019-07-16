package utils

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestWgetBasic(t *testing.T) {
	var conf Config
	conf.Url = "http://www.google.com"
	conf.Spoof = false
	conf.MaxErrors = 0
	conf.NoBackoff = false

	got, code, err := Wget(conf)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	want := "I'm Feeling Lucky"
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
	if code != 200 {
		t.Errorf("got %d; want %d", code, 200)
	}
}

func TestWgetSpoof(t *testing.T) {
	var conf Config
	conf.Url = "http://www.google.com"
	conf.Spoof = true
	conf.MaxErrors = 0
	conf.NoBackoff = false

	got, code, err := Wget(conf)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	want := "I'm Feeling Lucky"
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
	if code != 200 {
		t.Errorf("got %d; want %d", code, 200)
	}
}

func TestWgetErrors(t *testing.T) {
	maxReqsBefore503 := 20
	baseUrl := "https://www.boardgamegeek.com/xmlapi/boardgame/"

	var conf Config
	conf.Spoof = true
	conf.MaxErrors = 10
	conf.NoBackoff = false

	for i := 1; i <= maxReqsBefore503+10; i++ {
		conf.Url = fmt.Sprintf("%s%d", baseUrl, i)
		got, code, err := Wget(conf)
		if err != nil {
			t.Errorf("error: %s", err)
		}
		want := "<description>"
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}
		if code != 200 {
			t.Errorf("got %d; want %d", code, 200)
		}
	}
}

func TestWgetCode(t *testing.T) {
	var conf Config
	conf.Url = "https://httpstat.us/503"
	_, code, err := Wget(conf)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	if code != 503 {
		t.Errorf("got %d; want %d", code, 503)
	}
	conf.Url = "https://httpstat.us/202"
	_, code, err = Wget(conf)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	if code != 202 {
		t.Errorf("got %d; want %d", code, 202)
	}
}

func TestCache(t *testing.T) {
	now := time.Now()

	cache = make(map[string]cached)
	cache["www.google.com"] = cached{now, "I'm feeling lucky"}
	if err := saveCache("www.yahoo.com", "I'm feeling unlucky", 0); err != nil {
		t.Errorf("error: %s", err)
	}

	cache = make(map[string]cached)
	if err := loadCache(); err != nil {
		t.Errorf("error: %s", err)
	}

	if v, ok := cache["www.google.com"]; ok {
		{
			got := v.Expiry
			want := now
			if !got.Equal(want) {
				t.Errorf("got %s; want %s", got.String(), want.String())
			}
		}
		{
			got := v.Content
			want := "I'm feeling lucky"
			if got != want {
				t.Errorf("got %s; want %s", got, want)
			}
		}
	} else {
		t.Errorf("got %v; want %s", nil, "www.google.com")
	}

	if v, ok := cache["www.yahoo.com"]; ok {
		{
			got := v.Content
			want := "I'm feeling unlucky"
			if got != want {
				t.Errorf("got %s; want %s", got, want)
			}
		}
	} else {
		t.Errorf("got %v; want %s", nil, "www.yahoo.com")
	}
}

func TestExpiry(t *testing.T) {
	var conf Config
	conf.Url = "www.google.com"

	cache = make(map[string]cached)
	if err := saveCache(conf.Url, "I'm feeling lucky", time.Second*time.Duration(1)); err != nil {
		t.Errorf("error: %s", err)
	}

	cache = make(map[string]cached)
	if err := loadCache(); err != nil {
		t.Errorf("error: %s", err)
	}

	got := cacheGet(conf)
	want := "I'm feeling lucky"
	if got != want {
		t.Errorf("got %s; want %s", got, want)
	}

	time.Sleep(time.Second * time.Duration(1))

	got = cacheGet(conf)
	want = ""
	if got != want {
		t.Errorf("got %s; want %s", got, want)
	}
}
