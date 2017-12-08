package azkaban

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

type Client struct {
	SessionID     string
	http          *http.Client
	url           string
	DumpResponses bool
}

func (c *Client) ListFlows(project string) ([]Flow, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchprojectflows"
	params["project"] = project

	flows := ListFlowsResponse{}
	if err := c.requestAndDecode("GET", "manager", params, &flows); err != nil {
		return nil, err
	}

	return flows.Flows, nil
}

func (c *Client) ExecutionJobLog(executionID int64, jobID string) (string, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchExecJobLogs"
	params["execid"] = fmt.Sprintf("%d", executionID)
	params["jobId"] = jobID
	params["offset"] = "0"
	params["length"] = "10485760"

	log := FlowJobLog{}
	resp, err := c.request("GET", "executor", params)

	if err != nil {
		return "", err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&log)
	if err != nil {
		return "", err
	}

	return log.Data, nil
}

func (c *Client) FlowExecutions(project, flow string) (Executions, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchFlowExecutions"
	params["project"] = project
	params["flow"] = flow
	params["start"] = "0"
	params["length"] = "20"

	executions := ExecutionsList{}
	if err := c.requestAndDecode("GET", "manager", params, &executions); err != nil {
		return nil, err
	}

	return executions.Executions, nil
}

func (c *Client) FlowJobList(project, flow string) (FlowJobList, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchflowgraph"
	params["project"] = project
	params["flow"] = flow

	jobList := FlowJobList{}
	err := c.requestAndDecode("GET", "manager", params, &jobList)
	return jobList, err
}

func (c *Client) RestartFlowNow(project, flow string) error {
	params := make(map[string]string)
	params["ajax"] = "executeFlow"
	params["project"] = project
	params["flow"] = flow

	resp, err := c.request("GET", "executor", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Printf("%s", body)

	return nil
}

func (c *Client) FlowEcecutionStatus(executionID int64) (FlowExecutionStatus, error) {
	status := FlowExecutionStatus{}

	params := make(map[string]string)
	params["ajax"] = "fetchexecflow"
	params["execid"] = fmt.Sprintf("%d", executionID)

	resp, err := c.request("GET", "executor", params)
	if err != nil {
		return status, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&status)
	return status, err
}

func (c *Client) FlowSchedule(projectID int64, flowID string) (FlowSchedule, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchSchedule"
	params["projectId"] = fmt.Sprintf("%d", projectID)
	params["flowId"] = flowID

	res, err := c.request("GET", "schedule", params)
	if err != nil {
		return FlowSchedule{}, err
	}
	defer res.Body.Close()
	if c.DumpResponses {
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			return FlowSchedule{}, err
		}

		fmt.Printf("%s %s: \n", res.Request.Method, res.Request.URL.String())
		fmt.Printf("%s\n", b)
	}

	decoder := json.NewDecoder(res.Body)
	resp := ScheduleResponse{}
	err = decoder.Decode(&resp)
	if err != nil {
		return FlowSchedule{}, err
	}

	if resp.Empty() {
		return FlowSchedule{}, nil
	} else {
		return *resp.Schedule, err
	}
}

func (c *Client) requestAndDecode(method string, path string, params map[string]string, dst interface{}) error {
	resp, err := c.request(method, path, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if c.DumpResponses {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}

		fmt.Printf("%s %s: \n", method, resp.Request.URL.String())
		fmt.Printf("%s\n", b)
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(dst)
	if err != nil {
		return err
	}

	azkabanResp, ok := dst.(AzkabanError)
	if !ok {
		log.Printf("non-azkaban response for %s %s", method, resp.Request.URL)
		x, _ := httputil.DumpResponse(resp, true)
		log.Printf("%s", x)
		return fmt.Errorf("Bug: not an azakaban response")
	}
	if azkabanResp.AzkabanError() == "session" {
		return ErrInvalidSessionID
	}

	return nil

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
	return c.http.Do(req)
}