package main

import (
	"context"
	"encoding/json"
	"fmt"

	"bitbucket.org/mikehouston/asana-go"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/robfig/cron/v3"
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

// getAsanaClient retrieve token stored in database and then generates a google calendar service
func (p *Plugin) getAsanaClient(userID string) (*asana.Client, string) {
	var token oauth2.Token

	tokenInByte, appErr := p.API.KVGet(userID + "asanaToken")
	if appErr != nil {
		return nil, ""
	}

	json.Unmarshal(tokenInByte, &token)
	config := p.AsanaConfig()
	ctx := context.Background()
	tokenSource := config.TokenSource(ctx, &token)
	newToken, _ := tokenSource.Token()

	client := asana.NewClientWithAccessToken(newToken.AccessToken)

	return client, newToken.AccessToken
}

// CreateBotMessage used to post as asana bot to the user directly
func (p *Plugin) CreateBotMessage(userID, message string) *model.AppError {
	channel, err := p.API.GetDirectChannel(userID, p.botID)
	if err != nil {
		mlog.Error("Couldn't get bot's DM channel", mlog.String("user_id", userID))
		return err
	}

	post := &model.Post{
		UserId:    p.botID,
		ChannelId: channel.Id,
		Message:   message,
	}

	if _, err := p.API.CreatePost(post); err != nil {
		mlog.Error(err.Error())
		return err
	}

	return nil
}

func (p *Plugin) startCronJob(userID string) {
	cron := cron.New()
	cron.AddFunc("*/3 * * * *", func() {
		// p.CreateBotMessage(userID, "startCronJob call")
		p.GetNewEvents(userID)
	})
	cron.Start()
}
