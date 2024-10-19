package main

import (
	"ai-bot/internal/ai"
	"ai-bot/internal/data"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const (
	port = "8080"
)

func main() {
	slackSigningSecretContent, err := os.ReadFile("/etc/ai-chat-bot/slack-signing-secret")
	if err != nil {
		log.Fatal(err)
	}

	slackSigningSecret := string(slackSigningSecretContent)

	slackBotTokenContent, err := os.ReadFile("/etc/ai-chat-bot/slack-token")
	if err != nil {
		log.Fatal(err)
	}

	slackBotToken := string(slackBotTokenContent)

	threadRepo := data.NewThreadRepo()
	aiService, err := ai.NewService()
	api := slack.New(slackBotToken)

	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sv, err := slack.NewSecretsVerifier(r.Header, slackSigningSecret)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err := sv.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("%v\n", eventsAPIEvent)

		switch eventsAPIEvent.Type {
		case slackevents.URLVerification:
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			if _, err := w.Write([]byte(r.Challenge)); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case slackevents.CallbackEvent:
			innerEvent := eventsAPIEvent.InnerEvent

			log.Printf("%v\n", innerEvent)

			switch ev := innerEvent.Data.(type) {
			case *slackevents.MessageEvent:
				// filter other subtypes: https://api.slack.com/events/message#subtypes
				// log.Println("SubType: ", ev.SubType)
				// empty for me ^

				log.Println("TimeStamp: ", ev.TimeStamp, "Text: ", ev.Text, "ClientMsgID: ", ev.ClientMsgID, "ThreadTimeStamp: ", ev.ThreadTimeStamp)

				// if bot don't reply
				if ev.ClientMsgID == "" {
					return
				}

				// new message - create thread
				threadTimestamp := ev.TimeStamp
				prompt := ev.Text
				reply := true
				user := ev.Username
				history := []string{fmt.Sprintf("%s: %s", ev.Username, ev.Text)}

				if strings.Contains(strings.ToLower(ev.Text), "no_orc") {
					reply = false
				}

				// message in thread - reply in thread
				if ev.ThreadTimeStamp != "" {
					threadTimestamp = ev.ThreadTimeStamp
					messages, _, _, err := api.GetConversationReplies(&slack.GetConversationRepliesParameters{
						ChannelID: ev.Channel,
						Timestamp: ev.ThreadTimeStamp,
					})
					if err != nil {
						log.Println(err)
					}
					prompt = messages[len(messages)-1].Msg.Text
					user = messages[len(messages)-1].Msg.Username

					for i := range messages {
						if strings.Contains(strings.ToLower(messages[i].Text), "no_orc") {
							reply = false
						}
						history = append(history, fmt.Sprintf("%s: %s", messages[i].Username, messages[i].Text))
					}
				}

				if reply {
					go func() {
						thread := threadRepo.GetThread(r.Context(), ev.ThreadTimeStamp)
						output := aiService.GenerateFromKnowledge(thread, history, user, prompt)

						_, _, err = api.PostMessage(ev.Channel, slack.MsgOptionText(output.Text, false), slack.MsgOptionTS(threadTimestamp))
						if err != nil {
							log.Println(err)
						}

						err = threadRepo.SaveThread(r.Context(), thread)
						if err != nil {
							log.Println(err)
						}
					}()
				}
			}
		}
	})

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
