package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/mikehouston/asana-go"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"golang.org/x/oauth2"
)

// ServeHTTP allows the plugin to implement the http.Handler interface. Requests destined for the
// /plugins/{id} path will be routed to the plugin.
//
// The Mattermost-User-Id header will be present if (and only if) the request is by an
// authenticated user.
//
// This demo implementation sends back whether or not the plugin hooks are currently enabled. It
// is used by the web app to recover from a network reconnection and synchronize the state of the
// plugin's hooks.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch path := r.URL.Path; path {
	case "/oauth/connect":
		p.connectAsana(w, r)
	case "/oauth/complete":
		p.completeAsana(w, r)
	// case "/delete":
	// 	p.deleteEvent(w, r)
	// case "/handleresponse":
	// 	p.handleEventResponse(w, r)
	// case "/watch":
	// 	p.watchCalendar(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) connectAsana(w http.ResponseWriter, r *http.Request) {
	autheduserId := r.Header.Get("Mattermost-User-ID")

	if autheduserId == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	state := fmt.Sprintf("%v_%v", model.NewId()[10], autheduserId)

	if err := p.API.KVSet(state, []byte(state)); err != nil {
		http.Error(w, "Failed to save state", http.StatusBadRequest)
		return
	}

	asanaConfig := p.AsanaConfig()

	url := asanaConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (p *Plugin) completeAsana(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
		<head>
			<script>
				window.close();
			</script>
		</head>
		<body>
			<p>Completed connecting to Asana. Please close this window.</p>
		</body>
	</html>
	`
	autheduserId := r.Header.Get("Mattermost-User-ID")
	state := r.FormValue("state")
	code := r.FormValue("code")
	userId := strings.Split(state, "_")[1]
	config := p.AsanaConfig()
	if autheduserId == "" || userId != autheduserId {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	storedState, apiErr := p.API.KVGet(state)
	if apiErr != nil {
		http.Error(w, "Missing stored state", http.StatusBadRequest)
		return
	}

	if string(storedState) != state {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	if err := p.API.KVDelete(state); err != nil {
		http.Error(w, "Error deleting state", http.StatusBadRequest)
		return
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Error setting up Config Exchange", http.StatusBadRequest)
		return
	}

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		http.Error(w, "Invalid token marshal in completeAsana", http.StatusBadRequest)
		return
	}
	p.API.KVSet(userId+"asanaToken", tokenJSON)

	client := p.getAsanaClient(userId)

	workspaces, _, err := client.Workspaces(&asana.Options{Limit: 100})

	if err != nil || len(workspaces) == 0 {
		p.API.LogWarn("failed sync fresh calender", "error", err.Error())
		http.Error(w, "failed sync fresh calender", http.StatusInternalServerError)

		p.CreateBotMessage(userId, "Error while getting user workspaces")
		return
	}

	p.API.KVSet(userId+"asanaWorkspace", []byte(workspaces[0].ID))
	// TODO:

	// err = p.AsanaSync(userId)
	// if err != nil {
	// 	p.API.LogWarn("failed sync fresh calender", "error", err.Error())
	// 	http.Error(w, "failed sync fresh calender", http.StatusInternalServerError)
	// 	return
	// }

	// if err = p.setupCalendarWatch(userId); err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	// p.startCronJob(autheduserId)

	// // Post intro post
	// resJson, _ := json.Marshal(workspaces)

	message := "#### Welcome to the Mattermost Asana Plugin!\n" +
		"You've connected your account to your Asana.\n" +
		"Type **/asana help** to get bot usage help \n" +
		"Your `workspace` is `" + workspaces[0].Name + "`\n"

	if len(workspaces) > 1 {
		message += "Available workspases:\n"
		for i := 0; i < len(workspaces); i++ {
			message += "1. " + workspaces[i].Name + "\n"
		}
		message += "Workspace changing not implemented yet."
		// message += "If you want to change actual workspace, use `/asana workspace %ordinal number%` command` \n"
	}

	p.CreateBotMessage(userId, message)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// func (p *Plugin) deleteEvent(w http.ResponseWriter, r *http.Request) {
// 	html := `
// 	<!DOCTYPE html>
// 	<html>
// 		<head>
// 			<script>
// 				window.close();
// 			</script>
// 		</head>
// 	</html>
// 	`
// 	userId := r.Header.Get("Mattermost-User-ID")
// 	eventID := r.URL.Query().Get("evtid")
// 	calendarID := p.getPrimaryCalendarID(userId)
// 	srv, err := p.getAsanaClient(userId)
// 	if err != nil {
// 		p.CreateBotDMPost(userId, fmt.Sprintf("Unable to delete event. Error: %s", err))
// 		return
// 	}

// 	eventToBeDeleted, _ := srv.Events.Get(calendarID, eventID).Do()
// 	err = srv.Events.Delete(calendarID, eventID).Do()
// 	if err != nil {
// 		p.CreateBotDMPost(userId, fmt.Sprintf("Unable to delete event. Error: %s", err.Error()))
// 		return
// 	}

// 	p.CreateBotDMPost(userId, fmt.Sprintf("Success! Event _%s_ has been deleted.", eventToBeDeleted.Summary))
// 	w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 	fmt.Fprint(w, html)
// }

func (p *Plugin) handleEventResponse(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
		<head>
			<script>
				window.close();
			</script>
		</head>
	</html>
	`

	//TODO:

	// userId := r.Header.Get("Mattermost-User-ID")
	// response := r.URL.Query().Get("response")
	// eventID := r.URL.Query().Get("evtid")
	// calendarID := p.getPrimaryCalendarID(userId)
	// srv, _ := p.getAsanaClient(userId)

	// eventToBeUpdated, err := srv.Events.Get(calendarID, eventID).Do()
	// if err != nil {
	// 	p.CreateBotDMPost(userId, fmt.Sprintf("Error! Failed to update the response of _%s_ event.", eventToBeUpdated.Summary))
	// 	return
	// }

	// for idx, attendee := range eventToBeUpdated.Attendees {
	// 	if attendee.Self {
	// 		eventToBeUpdated.Attendees[idx].ResponseStatus = response
	// 	}
	// }

	// event, err := srv.Events.Update(calendarID, eventID, eventToBeUpdated).Do()
	// if err != nil {
	// 	p.CreateBotDMPost(userId, fmt.Sprintf("Error! Failed to update the response of _%s_ event.", event.Summary))
	// } else {
	// 	p.CreateBotDMPost(userId, fmt.Sprintf("Success! Event _%s_ response has been updated.", event.Summary))
	// }

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// func (p *Plugin) watchCalendar(w http.ResponseWriter, r *http.Request) {
// 	userId := r.URL.Query().Get("userId")
// 	channelID := r.Header.Get("X-Goog-Channel-ID")
// 	resourceID := r.Header.Get("X-Goog-Resource-ID")
// 	state := r.Header.Get("X-Goog-Resource-State")

// 	watchToken, _ := p.API.KVGet(userId + "watchToken")
// 	channelByte, _ := p.API.KVGet(userId + "watchChannel")
// 	var channel calendar.Channel
// 	json.Unmarshal(channelByte, &channel)
// 	if string(watchToken) == channelID && state == "exists" {
// 		p.AsanaSync(userId)
// 	} else {
// 		srv, _ := p.getAsanaClient(userId)
// 		srv.Channels.Stop(&calendar.Channel{
// 			Id:         channelID,
// 			ResourceId: resourceID,
// 		})
// 	}
// }
