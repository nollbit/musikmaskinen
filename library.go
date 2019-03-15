package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hajimehoshi/go-mp3"
	id3 "github.com/mikkyang/id3-go"
	log "github.com/sirupsen/logrus"
)

type (
	Track struct {
		Artist string
		Title  string
		Album  string
		Year   string
		Path   string
		Length int    // seconds
		Hash   string // sha1 string so we can use it as a key in the index map
	}

	// The index where we keep a content adressable cache
	Index map[string]*Track

	loadFromIndex func(string) (*Track, bool, error)

	// Library can add MP3 tracks from directories and keep an on-disk cache for quick startups
	Library struct {
		IndexPath string
		Tracks    []*Track
		Index     Index
	}
)

func getSha1(filename string) (string, error) {
	var s string

	file, err := os.Open(filename)
	if err != nil {
		return s, err
	}
	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return s, err
	}

	s = hex.EncodeToString(hash.Sum(nil))

	return s, nil

}

func getLength(filename string) (int, error) {

	f, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return 0, err
	}

	seconds := int(d.Length() / (int64(d.SampleRate()) * 2 * 2))

	defer d.Close()

	return seconds, nil
}

// extracts information about a single song
func extract(filename string, cacheLoader loadFromIndex) (*Track, error) {
	ctxLogger := log.WithField("filename", filename)
	hash, err := getSha1(filename)
	if err != nil {
		return nil, err
	}

	ctxLogger = ctxLogger.WithField("hash", hash)

	ctxLogger.Info("Reading file")

	song, found, err := cacheLoader(hash)
	if err != nil {
		return nil, err
	} else if found {
		ctxLogger.Info("Cache hit!")
		return song, nil
	}

	mp3File, err := id3.Open(filename)
	if err != nil {
		return nil, err
	}
	defer mp3File.Close()

	length, err := getLength(filename)
	if err != nil {
		return nil, err
	}

	song = &Track{
		Artist: strings.Trim(mp3File.Artist(), "\u0000"),
		Title:  strings.Trim(mp3File.Title(), "\u0000"),
		Album:  strings.Trim(mp3File.Album(), "\u0000"),
		Year:   strings.Trim(mp3File.Year(), "\u0000"),
		Path:   filename,
		Length: length,
		Hash:   hash,
	}

	return song, nil
}

func scan(rootPath string, cacheLoader loadFromIndex) ([]*Track, error) {
	songs := make([]*Track, 0)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".mp3" {
			return nil
		}

		song, err := extract(path, cacheLoader)
		if err != nil {
			return err
		}

		songs = append(songs, song)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return songs, nil
}

// create creates an index from a bunch of songs
func createIndex(songs []*Track) Index {
	i := make(map[string]*Track)

	for _, song := range songs {
		i[song.Hash] = song
	}

	return i
}

// serializeIndex takes an index and serializes it into a bunch of bytes
func serializeIndex(index Index) ([]byte, error) {
	jsonBytes, err := json.Marshal(index)

	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}

// unserializeIndex takes an index and serializes it into a bunch of bytes
func unserializeIndex(data []byte) (Index, error) {
	index := make(Index)
	err := json.Unmarshal(data, &index)

	if err != nil {
		return nil, err
	}

	return index, nil
}

func (l *Library) loadFromIndex(hash string) (*Track, bool, error) {
	song, ok := l.Index[hash]
	return song, ok, nil
}

// Add scans a directory and adds all songs found
func (l *Library) Add(rootPath string) error {
	songs, err := scan(rootPath, l.loadFromIndex)
	if err != nil {
		return err
	}

	l.Tracks = append(l.Tracks, songs...)

	return nil
}

// WriteIndex writes the index to disk. Run if after you've added all songs.
func (l *Library) WriteIndex() error {
	l.Index = createIndex(l.Tracks)

	indexBytes, err := serializeIndex(l.Index)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(l.IndexPath, indexBytes, 0644)
	return err
}

func (l *Library) readIndex() error {
	indexBytes, err := ioutil.ReadFile(l.IndexPath)
	if err != nil && os.IsNotExist(err) {
		// if the file doesn't exist, just create an empty index
		log.Info("No index, creating empty")
		l.Index = make(Index)
		return nil
	} else if err != nil {
		log.WithError(err).Error("Can't read index file")
		return err
	}

	index, err := unserializeIndex(indexBytes)
	if err != nil {
		log.WithError(err).Error("Can't unserialize index file")
		return err
	}
	log.Infof("Loaded index of size %d", len(index))

	l.Index = index

	return nil
}

// Sort sorts the song list in place. Run this after all libraries has been added.
func (l *Library) Sort() {
	sort.Slice(l.Tracks, func(i, j int) bool {
		if l.Tracks[i].Artist != l.Tracks[j].Artist {
			return l.Tracks[i].Artist < l.Tracks[j].Artist
		}
		return l.Tracks[i].Title < l.Tracks[j].Title
	})
}

// NewLibrary creates a new library and reads the index file.
func NewLibrary(indexPath string) (*Library, error) {
	library := &Library{Tracks: make([]*Track, 0), IndexPath: indexPath}
	err := library.readIndex()

	if err != nil {
		return nil, err
	}

	return library, nil
}
