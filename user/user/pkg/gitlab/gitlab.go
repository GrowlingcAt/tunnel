package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"net/url"
	"user/pkg/constants"
	"user/pkg/log"
)

type GitlabOAuth struct {
	domain       string
	clientID     string
	clientSecret string
	log          log.ILogger
}

func NewGitlabOAuth(domain, clientID, clientSecret string, log log.ILogger) *GitlabOAuth {
	return &GitlabOAuth{
		domain:       domain,
		clientID:     clientID,
		clientSecret: clientSecret,
		log:          log,
	}
}
func (gitlab *GitlabOAuth) getResourceUrl(apiPath string) string {
	return fmt.Sprintf("%s%s", gitlab.domain, apiPath)
}
func (gitlab *GitlabOAuth) GetOAuth2Config(redirectUrl string, params map[string]string, scopes []string) *oauth2.Config {
	u, _ := url.Parse(redirectUrl)
	ps, _ := url.ParseQuery(u.RawQuery)
	for k, v := range params {
		ps.Add(k, v)
	}
	u.RawQuery = ps.Encode()

	conf := &oauth2.Config{
		ClientID:     gitlab.clientID,
		ClientSecret: gitlab.clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  gitlab.getResourceUrl(constants.GITLABAUTHURL),
			TokenURL: gitlab.getResourceUrl(constants.GITLABTOKENURL),
		},
		RedirectURL: u.String(),
		Scopes:      scopes,
	}
	return conf
}

type GitlabUser struct {
	ID        int64  `json:"id"`
	UserName  string `json:"username"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
	Email     string `json:"email"`
}

func (gitlab *GitlabOAuth) GetUser(token *oauth2.Token) (*GitlabUser, error) {
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	res, err := client.Get(gitlab.getResourceUrl(constants.GITLABUSERURL))
	if err != nil {
		gitlab.log.Error(err)
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		gitlab.log.Error(err)
		return nil, err
	}
	u := &GitlabUser{}
	err = json.Unmarshal(body, u)
	if err != nil {
		gitlab.log.Error(err)
		return nil, err
	}
	return u, nil
}
