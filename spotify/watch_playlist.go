package spotify

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

const (
	pollInterval = 5 * time.Second
)

// WatchPlaylist subscribes to changes to a playlist
func WatchPlaylist(client *spotify.Client, playlistID spotify.ID, changes chan *spotify.FullPlaylist) error {
	log.Debugf("Setting up playlist watch for %s", playlistID)

	plLog := log.WithField("id", playlistID)

	go func() {
		// we can compare snapshot IDs to see if the playlist has changed.
		// setting it to empty string guarantees that we'll send out an initial message with the full playlist
		latestSnapshotID := ""

		for {
			plLog.Debug("Polling playlist")

			playlist, err := client.GetPlaylist(playlistID)

			if err != nil {
				plLog.WithError(err).Warn("Unable to poll playlist")
				continue
			}

			if playlist.SnapshotID != latestSnapshotID {
				changes <- playlist
				latestSnapshotID = playlist.SnapshotID
			}

			time.Sleep(pollInterval)
		}
	}()

	return nil
}
