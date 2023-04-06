package main

import (
	"github.com/mattermost/mattermost-plugin-api/experimental/command"
	"github.com/mattermost/mattermost-server/v5/model"
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
	cal := model.NewAutocompleteData("calendar", "[command]", "Available commands: connect, list, summary, create, help")

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
