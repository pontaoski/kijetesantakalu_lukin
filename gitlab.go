package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"
)

var client *gitlab.Client

var sort = "asc"
var tenpo = time.Minute / 100

func GitLab() {
	var err error
	client, err = gitlab.NewClient(config.GitLab.Token, gitlab.WithBaseURL(fmt.Sprintf("https://%s/api/v4", config.GitLab.URL)))
	if err != nil {
		panic(err)
	}
	println("GitLab handler is now running...")
outer:
	for {
		for _, group := range config.GitLab.Groups {
			projects, resp, err := client.Groups.ListGroupProjects(group, nil)
			if err != nil {
				fmt.Printf("%+v\n", err)
				fmt.Printf("%+v\n", resp)
				time.Sleep(tenpo)
				continue outer
			}
			for _, project := range projects {
				val := gitlab.PushedEventType
				events, resp, err := client.Events.ListProjectVisibleEvents(project.ID, &gitlab.ListContributionEventsOptions{Action: &val, Sort: &sort})
				if err != nil {
					fmt.Printf("%+v\n", err)
					fmt.Printf("%+v\n", resp)
					time.Sleep(tenpo)
					continue outer
				}
				var chats map[int64][]string = make(map[int64][]string)
				for _, chat := range GetChats() {
					var projects []string
					json.Unmarshal([]byte(chat.Data), &projects)
					chats[chat.ID] = projects
				}
			eventsLoop:
				for _, event := range events {
					if strings.HasPrefix(event.PushData.CommitTitle, "GIT_SILENT") ||
						strings.HasPrefix(event.PushData.CommitTitle, "SVN_SILENT") ||
						strings.HasPrefix(event.PushData.Ref, "work/") {
						continue eventsLoop
					}
					if event.PushData.CommitTitle == "" || event.PushData.CommitTo == "" {
						continue eventsLoop
					}
					if !TrackCommit(event.PushData.CommitTo, project.NameWithNamespace) {
						continue eventsLoop
					}
					data, keyboard := Embed{
						Title: EmbedHeader{
							Text: fmt.Sprintf("New commit from %s in %s", event.AuthorUsername, project.PathWithNamespace),
						},
						Fields: []EmbedField{
							{
								Title: "Title",
								Body:  event.PushData.CommitTitle,
							},
							{
								Title: "Branch",
								Body:  event.PushData.Ref,
							},
						},
					}, Keyboard{
						Title: "View on " + config.GitLab.URL,
						URL: fmt.Sprintf(
							"https://%s/%s/-/commit/%s",
							config.GitLab.URL,
							project.PathWithNamespace,
							event.PushData.CommitTo,
						),
					}
					telegramBot.SendMessage(config.Telegram.Channel, data, keyboard)
					for chat, projects := range chats {
						for _, subProject := range projects {
							if subProject == project.PathWithNamespace {
								telegramBot.SendMessage(chat, data, keyboard)
							}
						}
					}
				}
			}
		}
		time.Sleep(tenpo)
	}
}
