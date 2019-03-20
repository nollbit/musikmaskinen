package spotify

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/nollbit/spotify"
	"github.com/toqueteos/webbrowser"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	spotifyClientID     = kingpin.Flag("spotify-client-id", "Spotify client ID. See https://developer.spotify.com/dashboard/applications").Required().String()
	spotifyClientSecret = kingpin.Flag("spotify-client-secret", "Spotify client secret").Required().String()

	SpotifyCuratedPlaylistID = kingpin.
					Flag("spotify-curated-playlist", "The playlist from which people can select tracks. Must belong to the logged in user.").
					Default("5qpgQ7n2oPEw71GrLLnt85").
					String()

	oauthCallbackPort = kingpin.Flag("oauth-callback-port", "Where to redirect the user after login").Default("4040").Int()
)

func GetClient() (*spotify.Client, error) {

	auth := spotify.NewAuthenticator("http://localhost:4040/callback",
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopePlaylistModifyPrivate,
	)

	auth.SetAuthInfo(*spotifyClientID, *spotifyClientSecret)

	ch := make(chan *spotify.Client)
	state, err := state(32)

	if err != nil {
		return nil, err
	}

	// first start an HTTP server
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.Token(state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, state)
		}
		// use the token to get an authenticated client
		client := auth.NewClient(tok)
		fmt.Fprintf(w, "Login Completed!")
		ch <- &client
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":4040", nil)

	url := auth.AuthURL(state)

	webbrowser.Open(url)

	if err != nil {
		fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	}

	// wait for auth to complete
	client := <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	return client, nil
}

func state(n int) (string, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
