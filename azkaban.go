package azkabanlib

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func ConnectWithSessionID(url string, sessionID string) (*Client, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return &Client{
		http:      client,
		SessionID: sessionID,
		url:       url,
	}, nil
}

func ConnectWithUsernameAndPassword(url string, username string, password string) (*Client, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	q := req.URL.Query()
	q.Add("action", "login")
	q.Add("username", username)
	q.Add("password", password)
	req.URL.RawQuery = q.Encode()

	log.Printf("Authenticating against %s...", req.URL.String())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var status LoginResponse
	err = decoder.Decode(&status)
	if err != nil {
		return nil, err
	}

	if status.Status != "success" {
		return nil, fmt.Errorf("login failed")
	}

	return &Client{
		http:      client,
		SessionID: status.SessionID,
		url:       url,
	}, nil
}

type Client struct {
	SessionID string
	http      *http.Client
	url       string
}

func (c *Client) ProjectFlows(project string) ([]Flow, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchprojectflows"
	params["project"] = project

	resp, err := c.request("GET", "manager", params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	flows := ListFlowsResponse{}

	err = decoder.Decode(&flows)
	return flows.Flows, err
}

func (c *Client) ExecutionJobLog(executionID string, jobID string) (string, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchExecJobLogs"
	params["execid"] = executionID
	params["jobId"] = jobID
	params["offset"] = "0"
	params["length"] = "10485760"

	resp, err := c.request("GET", "executor", params)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	log := FlowJobLog{}

	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&log)
	if err != nil {
		return "", err
	}

	return log.Data, nil
}

func (c *Client) FlowExecutions(project, flow string) ([]Execution, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchFlowExecutions"
	params["project"] = project
	params["flow"] = flow
	params["start"] = "0"
	params["length"] = "20"

	resp, err := c.request("GET", "manager", params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	executions := ExecutionsList{}
	decoder.Decode(&executions)

	return executions.Executions, err
}

func (c *Client) FlowJobs(project, flow string) ([]FlowJob, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchflowgraph"
	params["project"] = project
	params["flow"] = flow

	resp, err := c.request("GET", "manager", params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	jobList := FlowJobList{}
	err = decoder.Decode(&jobList)
	if err != nil {
		return nil, err
	}

	return jobList.Nodes, nil
}

func (c *Client) request(method string, path string, params map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.url+path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("session.id", c.SessionID)
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	log.Printf("Requesting %s", req.URL.String())
	return c.http.Do(req)
}
