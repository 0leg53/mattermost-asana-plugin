// является костылём, ибо в
// https://bitbucket.org/mikehouston/asana-go
// нет реализации для Events
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"bitbucket.org/mikehouston/asana-go"
)

var ASANA_URL = "https://app.asana.com/api/1.0/"

type Event struct {
	Type string `json:"type"`
}

type Response struct {
	Sync    string        `json:"sync"`
	Errors  []asana.Error `json:"errors"`
	Data    []Event       `json:"data"`
	HasMore bool          `json:"has_more"`
}

func GetEventsResponse(accessToken string, project string, sync string) (Response, error) {
	url := ASANA_URL + "events"

	req, _ := http.NewRequest("GET", url, nil)

	q := req.URL.Query()
	q.Add("resource", project)
	q.Add("sync", sync)
	q.Add("opt_pretty", "true")
	q.Add("opt_feilds", "action,change,resource,type,user")

	req.URL.RawQuery = q.Encode()

	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", "Bearer "+accessToken)

	res, err := http.DefaultClient.Do(req)

	var result Response

	if err != nil {
		return result, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}
	return result, nil
}
