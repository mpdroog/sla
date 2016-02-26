package main

import (
	"testing"
	"time"
)

func TestMilliSeconds(t *testing.T) {
	df := map[string]float64{
		"3s": 3000,
		"10ms": 10,
		"1m": 1000*60,
	}
	for str, expect := range df {
		d, e := time.ParseDuration(str)
		if e != nil {
			t.Fatal(e)
		}
		ms := MilliSeconds(d)
		if ms != expect {
			t.Fatalf("Input(%s) expect=%d but got=%d", str, expect, ms)
		}
	}
}
