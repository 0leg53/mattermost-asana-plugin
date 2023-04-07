package main

import (
	"fmt"

	"golang.org/x/oauth2"
)

// AsanaConfig will return a oauth2 Config with the field set
func (p *Plugin) AsanaConfig() *oauth2.Config {
	config := p.API.GetConfig()
	clientID := p.getConfiguration().AsanaClientId
	clientSecret := p.getConfiguration().AsanaClientSecret

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.asana.com/-/oauth_authorize",
			TokenURL:  "https://app.asana.com/-/oauth_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: fmt.Sprintf("%s/plugins/%s/oauth/complete", *config.ServiceSettings.SiteURL, manifest.ID),
		Scopes:      nil,
		// []string{
		// 	"https://app.asana.com/-/oauth_authorize",
		// },
	}
}
