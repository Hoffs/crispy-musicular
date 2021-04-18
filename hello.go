package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify"
)

var (
	auth  = spotify.NewAuthenticator("http://localhost:3333/callback", spotify.ScopePlaylistReadPrivate)
	ch    = make(chan *spotify.Client)
	state = "test"
)

func main() {
	_ = godotenv.Load();
	_ = godotenv.Load(".env.local");

	auth.SetAuthInfo(os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"));
	var client *spotify.Client

	http.HandleFunc("/callback", redirectHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})

	go func() {
		url := auth.AuthURL(state)
		fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

		// wait for auth to complete
		client = <-ch

		// use the client to make calls that require authorization
		user, err := client.CurrentUser()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("You are logged in as:", user.ID)

		playlists, err := client.CurrentUsersPlaylists();
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Found %d playlists\n", playlists.Total)
		for i := 0; i < len(playlists.Playlists); i++ {
			p := playlists.Playlists[i];
			fmt.Printf("P: [%s] %s\n", p.ID, p.Name)
		}
		fmt.Printf("Has next? %s\n", playlists.Next);
	}()

	http.ListenAndServe(":3333", nil)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
}
