package main

import (
	"bitbucket.org/mikehouston/asana-go"
)

func (p *Plugin) GetTask(taskId string, client *asana.Client) *asana.Task {
	t := &asana.Task{
		ID: taskId,
	}

	t.Fetch(client)
	return t
}
