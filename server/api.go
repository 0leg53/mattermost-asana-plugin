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

	client, accessToken := p.getAsanaClient(userId)

	workspaces, _, err := client.Workspaces(&asana.Options{Limit: 100})

	if err != nil || len(workspaces) == 0 {
		errorMessage := "Error while getting user workspaces"
		p.API.LogWarn(errorMessage, "error", err.Error())
		http.Error(w, errorMessage, http.StatusInternalServerError)

		p.CreateBotMessage(userId, errorMessage)
		return
	}

	p.API.KVSet(userId+"asanaWorkspace", []byte(workspaces[0].ID))

	message := "#### Welcome to Asana Plugin!\n" +
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
	workspace := workspaces[0]
	// workspace.Projects(client)

	// TODO: remove, debug info
	workspaceJSON, err := json.Marshal(workspace)

	// TODO: remove, debug info
	message += "\n `" + string(workspaceJSON) + "` \n"

	Projects, _, _ := workspace.Projects(client)

	for _, project := range Projects {
		eventsResponse, _ := GetEventsResponse(accessToken, project.ID, "")
		p.API.KVSet(userId+":"+project.ID+"projectSyncToken", []byte(eventsResponse.Sync))
	}

	p.startCronJob(userId)

	p.CreateBotMessage(userId, message)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func (p *Plugin) GetNewEvents(userId string) {
	client, accessToken := p.getAsanaClient(userId)

	workspaceId, err := p.API.KVGet(userId + "asanaWorkspace")
	if err != nil {
		return
	}
	workspace := &asana.Workspace{
		ID: string(workspaceId),
	}
	Projects, _, _ := workspace.Projects(client)

	message := ""
	for _, project := range Projects {
		syncId, _ := p.API.KVGet(userId + ":" + project.ID + "projectSyncToken")
		eventsResponse, _ := GetEventsResponse(accessToken, project.ID, string(syncId))
		p.API.KVSet(userId+":"+project.ID+"projectSyncToken", []byte(eventsResponse.Sync))

		// TODO remove
		p.CreateBotMessage(userId, userId+":"+project.ID+"projectSyncToken")

		if len(eventsResponse.Errors) > 0 {
			eventsResponse, _ := GetEventsResponse(accessToken, project.ID, eventsResponse.Sync)
			p.API.KVSet(userId+":"+project.ID+"projectSyncToken", []byte(eventsResponse.Sync))
		}
		if len(eventsResponse.Data) > 0 {
			for index, event := range eventsResponse.Data {
				eventJSON, _ := json.Marshal(event)
				message += fmt.Sprintf("### Update %d \n", index+1) +
					"`" + string(eventJSON) + "`"
			}
		}

	}
	p.CreateBotMessage(userId, message)

}
