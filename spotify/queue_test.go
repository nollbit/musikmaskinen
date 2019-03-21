package spotify

/*
import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/go-test/deep"
)

func loadSongs(t *testing.T) []*Song {
	indexBytes, err := ioutil.ReadFile("testdata/songs.json")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	songs := make([]*Song, 0)
	err = json.Unmarshal(indexBytes, &songs)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	return songs
}

func TestQueue(t *testing.T) {
	songs := loadSongs(t)

	p := NewQueue(3)

	for _, s := range songs[0:3] {
		err := p.QueueAdd(s)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	err := p.QueueAdd(songs[4])
	if err != ErrorQueueFull {
		t.Errorf("Unexpected error %v", err)
		t.FailNow()
	}

	song2, err := p.QueueRemove()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	song1, err := p.QueueRemove()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	song0, err := p.QueueRemove()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if diff := deep.Equal(song0, songs[0]); diff != nil {
		t.Error(diff)
	}
	if diff := deep.Equal(song1, songs[1]); diff != nil {
		t.Error(diff)
	}
	if diff := deep.Equal(song2, songs[2]); diff != nil {
		t.Error(diff)
	}

	_, err = p.QueueRemove()
	if err != ErrorQueueEmpty {
		t.Errorf("Unexpected error %v", err)
		t.FailNow()
	}

}

func TestQueueNext(t *testing.T) {
	songs := loadSongs(t)

	p := NewQueue(3)

	for _, s := range songs[0:3] {
		err := p.QueueAdd(s)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	nextSong, err := p.Next()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if diff := deep.Equal(nextSong, songs[0]); diff != nil {
		t.Error(diff)
	}

	if len(p.queue) != 2 {
		t.Errorf("Expected queue to be len 2, but found %d", len(p.queue))
		t.FailNow()
	}

}
*/
