package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

var (
	auth         = getSpotifyAuthenticator()
	ch           = make(chan *spotify.Client)
	state        = uuid.New().String()
	redirectURI  = "http://localhost:8888/spotify-cli"
	clientId     = os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret = os.Getenv("SPOTIFY_SECRET")
)

type SpotifyAuthenticatorInterface interface {
	AuthURL(string) string
	Token(string, *http.Request) (*oauth2.Token, error)
	NewClient(*oauth2.Token) spotify.Client
}

func getSpotifyAuthenticator() SpotifyAuthenticatorInterface {
	auth := spotify.NewAuthenticator(
		redirectURI,
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserLibraryRead,
		// Used for Web Playback SDK
		"streaming",
		spotify.ScopeUserReadBirthdate,
		spotify.ScopeUserReadEmail,
	)
	auth.SetAuthInfo(clientId, clientSecret)
	return auth
}

// authenticate authenticate user with Sotify API
func authenticate() SpotifyClient {
	http.HandleFunc("/spotify-cli", authCallback)
	go http.ListenAndServe(":8888", nil)

	url := auth.AuthURL(state)

	err := openBroswerWith(url)
	if err != nil {
		log.Fatal(err)
	}

	client := <-ch

	token, _ := client.Token()
	t, _ := template.ParseFiles("index_tmpl.html")
	insertTokenToTemplate(token.AccessToken, t)
	return client
}

// openBrowserWith open browsers with given url
func openBroswerWith(url string) error {
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	default:
		err = fmt.Errorf("Sorry, %v OS is not supported", runtime.GOOS)
	}
	return err
}

// authCallback is a function to by Spotify upon successful
// user login at their site
func authCallback(w http.ResponseWriter, r *http.Request) {
	token, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusNotFound)
		return
	}

	client := auth.NewClient(token)
	ch <- &client

	user, err := getCurrentUser(client)
	fmt.Fprintf(w, "<h1>Logged into spotify cli as:</h1>\n<p>%v</p>", user.DisplayName)
}

var getCurrentUser = getCurrentUserWrapper

// getCurrentUserWrapper wraps CurrentUser method in order to be able to test it
func getCurrentUserWrapper(client spotify.Client) (*spotify.PrivateUser, error) {
	return client.CurrentUser()
}

type TemplateInterface interface {
	Execute(io.Writer, interface{}) error
}

type tokenToInsert struct {
	Token string
}

var osCreate = os.Create

func insertTokenToTemplate(token string, template TemplateInterface) error {
	f, err := osCreate("index.html")
	if err != nil {
		return err
	}
	defer f.Close()
	return template.Execute(f, tokenToInsert{token})
}