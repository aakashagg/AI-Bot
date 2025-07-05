package jira

import (
	"github.com/andygrunwald/go-jira"
)

// Service handles operations with Jira.
type Service struct {
	client     *jira.Client
	projectKey string
}

// NewService creates a Jira service using basic auth.
func NewService(baseURL, username, apiToken, projectKey string) (*Service, error) {
	tp := jira.BasicAuthTransport{
		Username: username,
		Password: apiToken,
	}
	client, err := jira.NewClient(tp.Client(), baseURL)
	if err != nil {
		return nil, err
	}
	return &Service{client: client, projectKey: projectKey}, nil
}

// CreateIssue creates a new issue in Jira and returns its key.
func (s *Service) CreateIssue(summary, description string) (string, error) {
	issue := &jira.Issue{
		Fields: &jira.IssueFields{
			Summary:     summary,
			Description: description,
			Type:        jira.IssueType{Name: "Task"},
			Project:     jira.Project{Key: s.projectKey},
		},
	}
	created, _, err := s.client.Issue.Create(issue)
	if err != nil {
		return "", err
	}
	return created.Key, nil
}

// IssueInfo returns the summary and description of the issue with the given key.
func (s *Service) IssueInfo(key string) (string, string, error) {
	issue, _, err := s.client.Issue.Get(key, nil)
	if err != nil {
		return "", "", err
	}
	return issue.Fields.Summary, issue.Fields.Description, nil
}
