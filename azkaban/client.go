package azkaban

import (
	"encoding/json"
	"errors"
	"fmt"
	htmlx "golang.org/x/net/html"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

type Client struct {
	SessionID     string
	http          *http.Client
	url           string
	DumpResponses bool
}

func (c *Client) ListProjects() ([]Project, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchallprojects"

	projects := ListAllProjectsResponse{}
	if err := c.requestAndDecode("GET", "index", params, &projects); err != nil {
		return nil, err
	}
	return projects.Projects, nil
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

// FetchLogsUntilEnd fetchs all logs from the given offset till the end and outputs them to the given writer
func (c *Client) FetchLogsUntilEnd(executionID int64, jobID string, offset int64, writer io.Writer) (int64, error) {
	fetchLength := int64(1024 * 512)
	currentOffset := offset
	for {
		log, err := c.FetchExecutionJobLog(executionID, jobID, currentOffset, fetchLength)
		if err != nil {
			return 0, err
		}

		writer.Write([]byte(html.UnescapeString(log.Data)))
		currentOffset += log.Length
		if log.Length == 0 || log.Length < fetchLength {
			break
		}
	}

	return currentOffset, nil
}

func (c *Client) FetchExecutionJobLog(executionID int64, jobID string, offset int64, length int64) (FlowJobLog, error) {
	params := make(map[string]string)
	params["ajax"] = "fetchExecJobLogs"
	params["execid"] = fmt.Sprintf("%d", executionID)
	params["jobId"] = jobID
	params["offset"] = fmt.Sprintf("%d", offset)
	params["length"] = fmt.Sprintf("%d", length)

	log := FlowJobLog{}
	resp, err := c.request("GET", "executor", params)

	if err != nil {
		return log, err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&log)
	return log, err
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

func (c *Client) FlowExecutionStatus(executionID int64) (FlowExecutionStatus, error) {
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

func (c *Client) Running() ([]FlowExecution, error) {
	resp, err := c.request("GET", "executor", nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("got %d %s when retrieving running executions", resp.StatusCode, resp.Status))
	}

	defer resp.Body.Close()
	doc, err := htmlx.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	// Azkaban serves the login page simply with a HTTP 200 so the only way to check if we're looking at the login page
	// is by looking for the login element.
	if findElementWithID(doc, "username") != nil && findElementWithID(doc, "password") != nil {
		return nil, errors.New("credentials expired, reauthenticate")
	}

	table := findElementWithID(doc, "executingJobs")

	return findExecutions(table)
}

type FlowExecution struct {
	FlowID  string
	Project string
	Execution
}

func findExecutions(n *htmlx.Node) ([]FlowExecution, error) {
	tbody := findElementsOfType(n, "tbody")
	rows := findElementsOfType(tbody[0], "tr")

	var executions []FlowExecution
	for _, row := range rows {
		cells := findElementsOfType(row, "td")

		execID, err := strconv.ParseInt(findElementsOfType(cells[1], "a")[0].FirstChild.Data, 10, 64)
		if err != nil {
			return nil, err
		}
		flowID := findElementsOfType(cells[3], "a")[0].FirstChild.Data
		project := findElementsOfType(cells[4], "a")[0].FirstChild.Data
		// TODO project ID is not part of the actual link so we can't parse it. We'll have to fetch projects before we do this
		// parsing time "2019-08-21 01:53:42" as "Jan 2, 2006 at 3:04pm": cannot parse "2019-08-21 01:53:42" as "Jan"
		startTime, err := time.Parse("2006-01-02 15:04:05", cells[7].FirstChild.Data)
		if err != nil {
			return nil, err
		}

		execution := FlowExecution{
			FlowID:  flowID,
			Project: project,
			Execution: Execution{
				SubmitTime: AzkabanTimestamp{},
				StartTime:  AzkabanTimestamp(startTime),
				Status:     "RUNNING",
				ID:         execID,
				EndTime:    AzkabanTimestamp(time.Now()),
			},
		}

		executions = append(executions, execution)
	}

	return executions, nil
}

func getAttribute(n *htmlx.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

func findElementsOfType(n *htmlx.Node, t string) []*htmlx.Node {
	var result []*htmlx.Node
	if n.Type == htmlx.ElementNode && n.Data == t {
		result = append(result, n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, findElementsOfType(c, t)...)
	}

	return result
}

func findElementWithID(n *htmlx.Node, id string) *htmlx.Node {
	if hasAttribute(n, "id", id) {
		return n;
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findElementWithID(c, id); result != nil {
			return result
		}
	}
	return nil
}

func hasAttribute(n *htmlx.Node, name string, value string) bool {
	for _, a := range n.Attr {
		if a.Key == name && a.Val == value {
			return true
		}
	}
	return false
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
	resp, err := c.http.Do(req)
	if c.DumpResponses {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}

		fmt.Printf("%s %s: \n", method, resp.Request.URL.String())
		fmt.Printf("%s\n", b)
	}

	return resp, err
}
