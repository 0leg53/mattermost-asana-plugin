package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-plugin-api/experimental/command"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

func (p *Plugin) getCommand() (*model.Command, error) {
	iconData, err := command.GetIconData(p.API, "assets/logo.svg")

	if err != nil {
		return nil, errors.Wrap(err, "failed to get icon data")
	}

	return &model.Command{
		Trigger:              "asana",
		DisplayName:          "Asana",
		Description:          "Integration with Asana",
		AutoComplete:         true,
		AutoCompleteDesc:     "Available commands: connect, list, summary, create, help",
		AutoCompleteHint:     "[command]",
		AutocompleteData:     getAutocompleteData(),
		AutocompleteIconData: iconData,
	}, nil
}

func getAutocompleteData() *model.AutocompleteData {
	cal := model.NewAutocompleteData("asana", "[command]", "Available commands: connect")

	connect := model.NewAutocompleteData("connect", "", "Connect your Asana with your Mattermost account")
	cal.AddCommand(connect)

	// list := model.NewAutocompleteData("list", "[number_of_events]", "List the upcoming X number of events")
	// list.AddTextArgument("Number of events to list", "[number_of_events]", "^[0-9]+$")
	// cal.AddCommand(list)

	// summary := model.NewAutocompleteData("summary", "[date]", "Get a breakdown of a particular date")
	// summary.AddTextArgument("The date to view in YYYY-MM-DD format", "[date]", "")
	// cal.AddCommand(summary)

	// create := model.NewAutocompleteData("create", "[title_of_event] [start_datetime] [end_datetime]", "Create an event with a title, start date-time and end date-time")
	// create.AddTextArgument("Title for the event you are creating, must be surrounded by quotes", "[title_of_event]", "")
	// create.AddTextArgument("Time the event starts in YYYY-MM-DD@HH:MM format", "[start_datetime]", "")
	// create.AddTextArgument("Time the event finishes in YYYY-MM-DD@HH:MM format", "[end_datetime]", "")
	// cal.AddCommand(create)

	// help := model.NewAutocompleteData("help", "", "Display usage")
	// cal.AddCommand(help)
	return cal
}

func (p *Plugin) postCommandResponse(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.botID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	_ = p.API.SendEphemeralPost(args.UserId, post)
}

// ExecuteCommand inside plugin
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split := strings.Fields(args.Command)
	command := split[0]
	action := ""
	config := p.API.GetConfig()

	if len(split) > 1 {
		action = split[1]
	}

	if command != "/asana" {
		return &model.CommandResponse{}, nil
	}

	if action == "connect" {
		if config.ServiceSettings.SiteURL == nil {
			p.postCommandResponse(args, "Invalid SiteURL")
			return &model.CommandResponse{}, nil
		} else {
			p.postCommandResponse(args, fmt.Sprintf("[Click here to link your Asana Account.](%s/plugins/%s/oauth/connect)", *config.ServiceSettings.SiteURL, manifest.ID))
			return &model.CommandResponse{}, nil
		}
	}
	messageToPost := ""
	// switch action {
	// case "list":
	// 	messageToPost = p.executeCommandList(args)
	// case "summary":
	// 	messageToPost = p.executeCommandSummary(args)
	// case "create":
	// 	messageToPost = p.executeCommandCreate(args)
	// case "help":
	// 	messageToPost = p.executeCommandHelp(args)
	// }

	if messageToPost != "" {
		p.postCommandResponse(args, messageToPost)
	}

	return &model.CommandResponse{}, nil
}
