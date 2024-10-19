package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type JiraScraper struct {
	urlBase           string
	username          string
	apiToken          string
	paginationOptions PaginationOptions
}

// Enter your project key in Jira here
const project = ""

type JiraScraperOptions struct {
	projectUrl        string
	fields            []string
	username          string
	apiToken          string
	paginationOptions PaginationOptions
}

type JiraResponse struct {
	Issues []Issue `json:"issues"`
}

type Issue struct {
	Key    string      `json:"key"`
	Fields IssueFields `json:"fields"`
}

type IssueFields struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

type PaginationOptions struct {
	start int
	max   int
	loops int
}

func NewJiraScraperFrom(options JiraScraperOptions) *JiraScraper {
	urlBase := options.projectUrl + "&fields=" + strings.Join(options.fields, ",")

	return &JiraScraper{
		urlBase:           urlBase,
		username:          options.username,
		apiToken:          options.apiToken,
		paginationOptions: options.paginationOptions,
	}
}

func (s *JiraScraper) Scrape() {
	getRequest := func(url string) *http.Request {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Accept", "application/json")
		req.SetBasicAuth(s.username, s.apiToken)
		return req
	}

	for i := 0; i < s.paginationOptions.loops; i++ {
		startAt := strconv.Itoa(s.paginationOptions.start + (i * s.paginationOptions.max))
		maxResults := strconv.Itoa(s.paginationOptions.max)

		url := s.urlBase + "&startAt=" + startAt + "&maxResults=" + maxResults
		request := getRequest(url)
		jiraResponse := GetJiraResponseFromRequest(request)
		jiraResponse.save()
	}
}

func main() {
	username := os.Getenv("ACQUIA_EMAIL")
	if username == "" {
		panic("username required (username@acquia.com)")
	}

	apiToken := os.Getenv("JIRA_TOKEN")
	if apiToken == "" {
		panic("jira api token required")
	}

	options := JiraScraperOptions{
		projectUrl: "https://atlassian.net/rest/api/2/search?jql=project=" + project,
		fields:     []string{"key", "summary", "description"},
		username:   username,
		apiToken:   apiToken,
		paginationOptions: PaginationOptions{
			start: 0,
			max:   2,
			loops: 2,
		},
	}

	scraper := NewJiraScraperFrom(options)

	scraper.Scrape()
}

func GetJiraResponseFromRequest(request *http.Request) *JiraResponse {
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var jiraResponse JiraResponse
	err = json.Unmarshal(body, &jiraResponse)
	if err != nil {
		log.Fatal(err)
	}

	return &jiraResponse
}

func (r *JiraResponse) save() {
	for i := range r.Issues {
		r.Issues[i].save()
	}
}

func (i *Issue) save() {
	fileName := "jira-tickets/" + i.Key + ".json"
	file, err := os.OpenFile(fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	issueAsJson, err := json.Marshal(i)
	if err != nil {
		log.Fatal(err)
	}

	file.Write(issueAsJson)
}
