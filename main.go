package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type SlackRequestBody struct {
	Text string `json:"text"`
}

type InboundSMS struct {
	Meta struct {
		Attempt      int64  `json:"attempt"`
		Delivered_to string `json:"delivered_to"`
	} `json:"meta"`

	Data struct {
		Event_type string `json:"event_type"`
		Id         string `json:"id"`
		Payload    struct {
			Text string `json:"text"`
			To   string `json:"to"`
		} `json:"payload"`
	} `json:"data"`
}

var webhookUrl string = "https://hooks.slack.com/services/T04XXXXXX/B01YYYYYYYYFJ/Dbz4JZZZZZZZZZZU1DEIVM"

func SMSHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Fatal(err)
	}

	var request InboundSMS
	err = json.Unmarshal([]byte(bodyBytes), &request)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Delivered to Webhook: %v\n", request.Meta.Delivered_to)
	fmt.Printf("Event Type: %v\n", request.Data.Event_type)
	fmt.Printf("ID: %v\n", request.Data.Id)
	fmt.Printf("To: %v\n", request.Data.Payload.To)
	fmt.Printf("Inbound Message:\n%v\n", request.Data.Payload.Text)

	var message string = "Text Message:\n" + request.Data.Payload.Text
	SendSlackNotification(webhookUrl, message)

}

func Healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func SendSlackNotification(webhookUrl string, msg string) error {

	slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
	req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", SMSHandler).Methods("POST")
	r.HandleFunc("/", Healthz).Methods("GET")
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	srv := &http.Server{
		Handler:      loggedRouter,
		Addr:         "0.0.0.0:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Starting server on address", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
