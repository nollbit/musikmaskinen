package main

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/go-test/deep"
)

func TestScan(t *testing.T) {

	cacheRequests := make([]string, 0)
	cacheLoader := func(hash string) (*Song, bool, error) {
		cacheRequests = append(cacheRequests, hash)
		return nil, false, nil
	}

	songs, err := scan("./testdata", cacheLoader)

	if err != nil {
		t.Error(err)
		return
	}

	expectedKeys := []string{
		"e5f9829dc96c222a2140d87de918383b8cea81d3",
		"0052549a811876dff2ed75be1d9f2ff730d09267",
		"7d7e4e1a8f83787db9919979b390d0f5da70229b",
	}

	if diff := deep.Equal(sortStringSlice(cacheRequests), sortStringSlice(expectedKeys)); diff != nil {
		t.Error(diff)
	}

	if len(songs) != 3 {
		t.Errorf("Got %d songs, expected 4", len(songs))
	}
}

func j(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}

	return string(b)
}

func loadTestSongs(t *testing.T) []*Song {
	songs, err := scan("./testdata", func(_ string) (*Song, bool, error) {
		return nil, false, nil
	})

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	return songs
}

func TestExtract(t *testing.T) {

	cacheLoader := func(hash string) (*Song, bool, error) {
		// make sure we get the hash we expect
		if hash != "7d7e4e1a8f83787db9919979b390d0f5da70229b" {
			t.Errorf("Expected another hash than %s", hash)
			t.FailNow()
		}
		return nil, false, nil
	}

	song, err := extract("testdata/David_Szesztay_-_Cheese.mp3", cacheLoader)

	if err != nil {
		t.Error(err)
		return
	}

	expected := &Song{
		Artist: "David Szesztay",
		Title:  "Cheese",
		Album:  "Commercial",
		Year:   "",
		Path:   "testdata/David_Szesztay_-_Cheese.mp3",
		Length: 31,
		Hash:   "7d7e4e1a8f83787db9919979b390d0f5da70229b",
	}

	if diff := deep.Equal(song, expected); diff != nil {
		t.Error(diff)
	}

}

func TestExtractWithCacheHit(t *testing.T) {
	cachedSong := &Song{
		Artist: "David Szesztay (CACHED)",
		Title:  "Cheese",
		Album:  "Commercial",
		Year:   "",
		Path:   "testdata/David_Szesztay_-_Cheese.mp3",
		Length: 31,
		Hash:   "7d7e4e1a8f83787db9919979b390d0f5da70229b",
	}

	cacheLoader := func(hash string) (*Song, bool, error) {
		// make sure we get the hash we expect
		if hash != "7d7e4e1a8f83787db9919979b390d0f5da70229b" {
			t.Errorf("Expected another hash than %s", hash)
			t.FailNow()
		}
		return cachedSong, true, nil
	}

	song, err := extract("testdata/David_Szesztay_-_Cheese.mp3", cacheLoader)

	if err != nil {
		t.Error(err)
		return
	}

	if diff := deep.Equal(song, cachedSong); diff != nil {
		t.Error(diff)
	}

}
func sortStringSlice(s []string) []string {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})

	return s
}

func TestCreateIndex(t *testing.T) {
	songs := loadTestSongs(t)

	index := createIndex(songs)

	expectedKeys := []string{
		"e5f9829dc96c222a2140d87de918383b8cea81d3",
		"0052549a811876dff2ed75be1d9f2ff730d09267",
		"7d7e4e1a8f83787db9919979b390d0f5da70229b",
	}

	keys := make([]string, 0, len(index))
	for k := range index {
		keys = append(keys, k)
	}

	if diff := deep.Equal(sortStringSlice(keys), sortStringSlice(expectedKeys)); diff != nil {
		t.Error(diff)
	}

	s1, ok := index["7d7e4e1a8f83787db9919979b390d0f5da70229b"]
	if !ok {
		t.Error("Expected to find a certain key in index, but didn't find it")
		return
	}

	expected := &Song{
		Artist: "David Szesztay",
		Title:  "Cheese",
		Album:  "Commercial",
		Year:   "",
		Path:   "testdata/David_Szesztay_-_Cheese.mp3",
		Length: 31,
		Hash:   "7d7e4e1a8f83787db9919979b390d0f5da70229b",
	}

	if diff := deep.Equal(s1, expected); diff != nil {
		t.Error(diff)
	}
}

func TestIndexSerialization(t *testing.T) {
	songs := loadTestSongs(t)

	index := createIndex(songs)

	serialized, err := serializeIndex(index)

	if err != nil {
		t.Error(err)
		return
	}

	index2, err := unserializeIndex(serialized)

	if diff := deep.Equal(index, index2); diff != nil {
		t.Error(diff)
	}
}
