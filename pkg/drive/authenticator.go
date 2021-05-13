package drive

import (
	"context"
	"errors"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type auth struct {
	config oauth2.Config
}

type Authenticator interface {
	AuthURL() string
	Token(r *http.Request) (*oauth2.Token, error)
	NewClient(token *oauth2.Token) (*drive.Service, error)
	FromRefreshToken(token string) (*drive.Service, error)
}

func NewAuthenticator(id string, secret string, redirectURL string) Authenticator {
	return &auth{
		config: oauth2.Config{
			ClientID:     id,
			ClientSecret: secret,
			RedirectURL:  redirectURL,
			Scopes:       []string{drive.DriveFileScope},
			Endpoint:     google.Endpoint,
		},
	}

}

func (a *auth) AuthURL() string {
	return a.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

func (a *auth) Token(r *http.Request) (*oauth2.Token, error) {
	values := r.URL.Query()

	if e := values.Get("error"); e != "" {
		return nil, errors.New("drive: auth failed - " + e)
	}
	code := values.Get("code")
	if code == "" {
		return nil, errors.New("drive: didn't get access code")
	}

	return a.config.Exchange(context.Background(), code)
}

func (a *auth) NewClient(token *oauth2.Token) (*drive.Service, error) {
	ctx := context.Background()
	return drive.NewService(ctx, option.WithTokenSource(a.config.TokenSource(ctx, token)))
}

func (a *auth) FromRefreshToken(token string) (*drive.Service, error) {
	ctx := context.Background()
	oauth := &oauth2.Token{RefreshToken: token}

	return drive.NewService(ctx, option.WithTokenSource(a.config.TokenSource(ctx, oauth)))
}
