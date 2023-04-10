package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"bitbucket.org/mikehouston/asana-go"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
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

// AsanaSync either does a full sync or a incremental sync.
// Taken from googles sample code
// To better understand whats going on here, you can read https://developers.google.com/calendar/v3/sync
func (p *Plugin) AsanaSync(userID string) error {
	client, _ := p.getAsanaClient(userID)

	if client == nil {
		return errors.New("cant't get asana client")
	}
	// TODO:

	// request := client.Events.List("primary")

	// isIncrementalSync := false
	// syncTokenByte, KVGetErr := p.API.KVGet(userID + "syncToken")
	// syncToken := string(syncTokenByte)
	// if KVGetErr != nil || syncToken == "" {
	// 	// Perform a Full Sync
	// 	sixMonthsFromNow := time.Now().AddDate(0, 6, 0).Format(time.RFC3339)
	// 	request.TimeMin(time.Now().Format(time.RFC3339)).TimeMax(sixMonthsFromNow).SingleEvents(true)
	// } else {
	// 	// Performing a Incremental Sync
	// 	request.SyncToken(syncToken).ShowDeleted(true)
	// 	isIncrementalSync = true
	// }

	// var pageToken string
	// var events *calendar.Events
	// var allEvents []*calendar.Event
	// for ok := true; ok; ok = pageToken != "" {
	// 	request.PageToken(pageToken)
	// 	events, err = request.Do()
	// 	if err != nil {
	// 		p.API.KVDelete(userID + "syncToken")
	// 		p.API.KVDelete(userID + "events")
	// 		p.AsanaSync(userID)
	// 	}
	// 	if len(events.Items) != 0 {
	// 		for _, item := range events.Items {
	// 			allEvents = append(allEvents, item)
	// 		}
	// 	}
	// 	pageToken = events.NextPageToken
	// }

	// p.API.KVSet(userID+"syncToken", []byte(events.NextSyncToken))
	// if !isIncrementalSync {
	// 	sort.Slice(allEvents, func(i, j int) bool {
	// 		return allEvents[i].Start.DateTime < allEvents[j].Start.DateTime
	// 	})
	// 	allEventsJSON, _ := json.Marshal(allEvents)
	// 	p.API.KVSet(userID+"events", allEventsJSON)
	// } else {
	// 	p.updateEventsInDatabase(userID, allEvents)
	// }

	return nil
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
