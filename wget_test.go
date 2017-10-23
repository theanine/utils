package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestWgetBasic(t *testing.T) {
	var conf Config
	conf.Url = "http://www.google.com"
	conf.Spoof = false
	conf.MaxErrors = 0
	conf.NoBackoff = false

	got := Wget(conf)
	want := "I'm Feeling Lucky"
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestWgetSpoof(t *testing.T) {
	var conf Config
	conf.Url = "http://www.google.com"
	conf.Spoof = true
	conf.MaxErrors = 0
	conf.NoBackoff = false

	got := Wget(conf)
	want := "I'm Feeling Lucky"
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestWgetErrors(t *testing.T) {
	maxReqsBefore503 := 20
	baseUrl := "https://www.boardgamegeek.com/xmlapi/boardgame/"

	var conf Config
	conf.Spoof = true
	conf.MaxErrors = 10
	conf.NoBackoff = false

	for i := 1; i <= maxReqsBefore503*1.5; i++ {
		conf.Url = fmt.Sprintf("%s%d", baseUrl, i)
		got := Wget(conf)
		want := "<description>"
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", want)
		}
	}
}
