package github

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"user/pkg/log"

	"golang.org/x/oauth2"
)

const (
	GITHUBAUTHURL  = "https://github.com/login/oauth/authorize"
	GITHUBTOKENURL = "https://github.com/login/oauth/access_token"
	GITHUBUSERURL  = "https://api.github.com/user"
)

type GithubOAuth struct {
	clientID     string
	clientSecret string
	log          log.ILogger
}

func NewGithubOAuth(clientID, clientSecret string, log log.ILogger) *GithubOAuth {
	return &GithubOAuth{
		clientID:     clientID,
		clientSecret: clientSecret,
		log:          log,
	}
}

func (github *GithubOAuth) GetOAuth2Config(redirectURL string, params map[string]string, scopes []string) *oauth2.Config {
	u, _ := url.Parse(redirectURL)
	ps, _ := url.ParseQuery(u.RawQuery)

	for k, v := range params {
		ps.Add(k, v)
	}

	u.RawQuery = ps.Encode()

	return &oauth2.Config{
		ClientID:     github.clientID,
		ClientSecret: github.clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  GITHUBAUTHURL,
			TokenURL: GITHUBTOKENURL,
		},
		RedirectURL: u.String(),
		Scopes:      scopes,
	}
}

type GithubUser struct {
	ID        int64  `json:"id"`
	UserName  string `json:"login"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
	Email     string `json:"email"`
}

func (github *GithubOAuth) GetUser(token *oauth2.Token) (*GithubUser, error) {
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))

	res, err := client.Get(GITHUBUSERURL)
	if err != nil {
		github.log.Error(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		github.log.Error(err)
		return nil, err
	}

	user := &GithubUser{}
	err = json.Unmarshal(body, user)
	if err != nil {
		github.log.Error(err)
		return nil, err
	}

	return user, nil
}
