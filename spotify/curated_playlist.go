package spotify

import (
	"sort"
	"strings"
	"time"

	"github.com/nollbit/spotify"
	log "github.com/sirupsen/logrus"
)

type CuratedPlaylist struct {
	PlaylistID spotify.ID
	Tracks     []spotify.FullTrack
	Changes    chan string              // is there a no-op type for channels?
	blacklist  map[spotify.ID]time.Time // stores the track blacklist
}

func (c *CuratedPlaylist) BlacklistTrack(trackID spotify.ID, duration time.Duration) {
	c.blacklist[trackID] = time.Now().Add(duration)
	log.Debugf("blacklisted now until %s", time.Now().Add(duration))

}

func (c *CuratedPlaylist) IsTrackBlacklisted(trackID spotify.ID) (time.Time, bool) {
	t, ok := c.blacklist[trackID]

	if ok && t.After(time.Now()) {
		return t, true
	}

	return time.Now(), false

}

func NewCuratedPlaylist(spotifyClient *spotify.Client, playlistID spotify.ID) (*CuratedPlaylist, error) {
	log.Debugf("Creating curated playlist from %s", playlistID)

	curatedPlaylistChanges := make(chan *spotify.FullPlaylist)
	err := WatchPlaylist(spotifyClient, playlistID, curatedPlaylistChanges)
	if err != nil {
		log.WithError(err).Error("Unable to watch playlist")
		return nil, err
	}

	c := &CuratedPlaylist{
		PlaylistID: playlistID,
		Tracks:     make([]spotify.FullTrack, 0),
		Changes:    make(chan string),
		blacklist:  make(map[spotify.ID]time.Time),
	}

	go func() {
		for {
			curatedPlaylist := <-curatedPlaylistChanges

			newCuratedPlaylistTracks := make([]spotify.FullTrack, 0, len(c.Tracks))

			// So, the spotify pkg doesn't page, so we'll have to do that manually for now
			// https://github.com/zmb3/spotify/pull/79
			page := &curatedPlaylist.Tracks

			morePages := true
			for morePages {
				plTracks := page.Tracks
				log.Debugf("Adding %d tracks from page", len(page.Tracks))
				for _, plTrack := range plTracks {
					newCuratedPlaylistTracks = append(newCuratedPlaylistTracks, plTrack.Track)
				}

				var err error
				if page.Next != "" {
					newPage := &spotify.PlaylistTrackPage{}
					err = spotifyClient.Get(page.Next, newPage)

					if err != nil {
						log.WithError(err).Error("Unable to get next page")
						break
					}

					page = newPage
				} else {
					break
				}

			}

			if err != nil {
				// we really don't want to die, so let's just ignore this round of changes instead
				continue
			}

			// sort by first artist name (case insensitive), track name desc
			sort.Slice(newCuratedPlaylistTracks, func(i, j int) bool {
				artistI := strings.ToLower(newCuratedPlaylistTracks[i].Artists[0].Name)
				artistJ := strings.ToLower(newCuratedPlaylistTracks[j].Artists[0].Name)
				if artistI != artistJ {
					return artistI < artistJ
				}
				return newCuratedPlaylistTracks[i].Name < newCuratedPlaylistTracks[j].Name
			})

			c.Tracks = newCuratedPlaylistTracks
			c.Changes <- curatedPlaylist.SnapshotID
		}
	}()

	return c, nil
}
