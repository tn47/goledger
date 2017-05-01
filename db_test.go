package main

import "testing"
import "time"
import "fmt"
import "sort"
import "math/rand"
import "reflect"

func TestDBInsert(t *testing.T) {
	for n := 0; n < 10; n++ {
		t.Logf("case %v", n)

		times := []time.Time{}
		for i := 0; i < n; i++ {
			if len(times) == 0 || (rand.Intn(20000)%2) == 0 {
				times = append(times, time.Now())
				time.Sleep(1 * time.Millisecond)
			} else {
				times = append(times, times[len(times)-1])
			}
		}

		ref := []string{}
		db := NewDB("testing")
		for i := range times {
			k := times[rand.Intn(len(times))]
			db.Insert(k, i)
			ref = append(ref, fmt.Sprintf("%v:%v", k, i))
		}

		sort.Strings(ref)

		results := []string{}
		entries := db.Range(nil, nil, "both", make([]KV, 0))
		for _, entry := range entries {
			results = append(results, fmt.Sprintf("%v:%v", entry.k, entry.v))
		}

		sort.Strings(results)
		if reflect.DeepEqual(results, ref) == false {
			t.Fatalf("expected %v, got %v", ref, results)
		}
	}

}

func TestDBRange(t *testing.T) {
	verify := func(reftimes []time.Time, entries []KV) {
		refs := []string{}
		for _, tm := range reftimes {
			refs = append(refs, fmt.Sprintf("%v", tm))
		}
		results := []string{}
		for _, entry := range entries {
			results = append(results, fmt.Sprintf("%v", entry.k))
		}
		if reflect.DeepEqual(results, refs) == false {
			t.Fatalf("expected %v, got %v", refs, results)
		}
	}

	for n := 1; n < 10; n++ {
		t.Logf("case %v", n)

		db := NewDB("testing")
		times := []time.Time{}
		for i := 0; i < n; i++ {
			time.Sleep(1 * time.Millisecond)
			tm := time.Now()
			db.Insert(tm, i)
			times = append(times, tm)
		}

		for i := 0; i < len(times)-2; i++ {
			for j := i; j < len(times)-1; j++ {
				low, high := times[i], times[j]
				t.Logf("%v:%v %v:%v", i, low, j, high)
				if i+1 <= j-1 {
					entries := db.Range(&low, &high, "none", make([]KV, 0))
					verify(times[i+1:j], entries)
					entries = db.Range(&low, &high, "low", make([]KV, 0))
					verify(times[i:j], entries)
					entries = db.Range(&low, &high, "high", make([]KV, 0))
					verify(times[i+1:j+1], entries)
				}
				entries := db.Range(&low, &high, "both", make([]KV, 0))
				verify(times[i:j+1], entries)
			}
		}
	}
}
