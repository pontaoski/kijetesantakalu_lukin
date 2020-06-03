package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"
	"golang.org/x/time/rate"
)

var client *gitlab.Client

var sort = "asc"
var tenpo = time.Minute / 100

type Event struct {
	Group   string
	Project *gitlab.Project
	*gitlab.ContributionEvent
}

var eventsChan = make(chan Event, 100)
var limiter = rate.NewLimiter(rate.Every(tenpo), 100)
var ctx = context.Background()

func GitlabReader() {
	sort := "desc"
	eType := gitlab.PushedEventType
	for _, group := range allGroups {
		go func(group string) {
			limiter.Wait(ctx)
			projects, resp, err := client.Groups.ListGroupProjects(group, nil)
			if err != nil {
				fmt.Printf("%+v\n", err)
				fmt.Printf("%+v\n", resp)
				return
			}
			for _, project := range projects {
				go func(project *gitlab.Project) {
					limiter.Wait(ctx)
					events, resp, err := client.Events.ListProjectVisibleEvents(project.ID, &gitlab.ListContributionEventsOptions{Action: &eType, Sort: &sort})
					if err != nil {
						fmt.Printf("%+v\n", err)
						fmt.Printf("%+v\n", resp)
						return
					}
					for _, event := range events {
						eventsChan <- Event{group, project, event}
					}
				}(project)
			}
		}(group)
	}
}

func GitLab() {
	var err error
	client, err = gitlab.NewClient(config.GitLab.Token, gitlab.WithBaseURL(fmt.Sprintf("https://%s/api/v4", config.GitLab.URL)))
	if err != nil {
		panic(err)
	}
	println("GitLab handler is now running...")
	go GitlabReader()
	for {
		select {
		case event := <-eventsChan:
			if strings.HasPrefix(event.PushData.CommitTitle, "GIT_SILENT") ||
				strings.HasPrefix(event.PushData.CommitTitle, "SVN_SILENT") ||
				strings.HasPrefix(event.PushData.Ref, "work/") {
				continue
			}
			if event.PushData.CommitTitle == "" || event.PushData.CommitTo == "" {
				continue
			}
			if !TrackCommit(event.PushData.CommitTo, event.Project.NameWithNamespace) {
				continue
			}
			data, keyboard := Embed{
				Title: EmbedHeader{
					Text: fmt.Sprintf("New commit from %s in %s", event.AuthorUsername, event.Project.NameWithNamespace),
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
					event.Project.NameWithNamespace,
					event.PushData.CommitTo,
				),
			}
			for _, channel := range config.Telegram.Channels {
				if Contains(channel.Groups, event.Group) || Contains(channel.Projects, event.Project.NameWithNamespace) {
					telegramBot.SendMessage(channel.ID, data, keyboard)
				}
			}
			var chatGroups map[int64][]string = make(map[int64][]string)
			var chatProjects map[int64][]string = make(map[int64][]string)
			for _, chat := range GetChats() {
				var projects []string
				var groups []string
				json.Unmarshal([]byte(chat.ProjectData), &projects)
				chatProjects[chat.ID] = projects
				json.Unmarshal([]byte(chat.GroupData), &groups)
				chatGroups[chat.ID] = groups
			}
			for chat, groups := range chatGroups {
				for _, subGroup := range groups {
					if subGroup == event.Group {
						telegramBot.SendMessage(chat, data, keyboard)
					}
				}
			}
			for chat, projects := range chatProjects {
				for _, subProject := range projects {
					if subProject == event.Project.PathWithNamespace {
						telegramBot.SendMessage(chat, data, keyboard)
					}
				}
			}
		}
		time.Sleep(tenpo)
	}
}
