package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

func newCheckFlowCmd(context Context) *cobra.Command {
	return &cobra.Command{
		Use:     "flow",
		Aliases: []string{"f"},
		Short:   "check a given flow",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			flowRepo := context.Context().Flows()
			flow, proj, err := flowRepo.Flow(azkaban.Project{Name: context.Project()}, args[0])
			if err != nil {
				log.Fatal(err)
			}
			statusChecker := FlowStatusChecker{
				client:         context.Client(),
				context:        context.Context(),
				histogramCount: 6,
				project:        proj,
				flow:           flow,
			}

			status, err := statusChecker.printFlowStatus()
			if err != nil {
				log.Fatal(err)
			}

			if status.Health == azkaban.Critical {
				fmt.Printf("Execution failed in %q, log messages of interest:\n", status.FailedJob.ID)
				client := context.Client()
				buffer := bytes.NewBuffer([]byte{})
				_, err := client.FetchLogsUntilEnd(status.LastExecution.ID, status.FailedJob.ID, 0, buffer)
				if err != nil {
					log.Fatal(err)
				}

				scanner := bufio.NewScanner(strings.NewReader(buffer.String()))

				patterns := []string{
					"err",
					"exception",
					"failed",
					"failure",
				}
				fmt.Println(strings.Repeat("-", 80))
				for scanner.Scan() {
					line := scanner.Text()
					lower := strings.ToLower(line)
					ofInterest := false
					for _, p := range patterns {
						if strings.Contains(lower, p) {
							ofInterest = true
							break
						}
					}
					if ofInterest {
						fmt.Println(line)
					}
				}
				fmt.Println(strings.Repeat("-", 80))
				fmt.Println()

				scanner = bufio.NewScanner(os.Stdin)
				run := true

				printActionTerm()
				for run {
					printActionTerm()
					run = run && scanner.Scan()
					input := strings.ToLower(strings.TrimSpace(scanner.Text()))

					if input == "restart" {
						client.RestartFlowNow(proj.Name, flow.FlowID)
					} else if input == "logs" {
						// TODO this might be slow:
						// fmt.Println(l)
						_, err = client.FetchLogsUntilEnd(status.LastExecution.ID, status.FailedJob.ID, 0, os.Stdout)
						if err != nil {
							log.Fatal(err)
						}
					} else if input == "status" {
						status, err = statusChecker.printFlowStatus()
						if err != nil {
							log.Fatal(err)
						}
					} else if input == "unschedule" {
						fmt.Println("not implemented yet")
					} else {
						run = false
					}
				}
			}
		},
	}
}

type FlowStatusChecker struct {
	client         *azkaban.Client
	context        *azkaban.Context
	histogramCount uint
	project        azkaban.Project
	flow           azkaban.Flow
}

func (h FlowStatusChecker) printSchedule() error {
	client := h.client
	schedule, err := client.FlowSchedule(h.project.ID, h.flow.FlowID)
	if err != nil {
		return err
	}
	scheduledMessage := "not scheduled"
	if schedule.IsScheduled() {
		scheduledMessage = fmt.Sprintf("%s", humanize.Time(schedule.NextExecTime.Time()))
	}
	fmt.Printf("%-16s %s\n", "Next execution:", scheduledMessage)
	return nil
}

func (h FlowStatusChecker) printFlowStatus() (status FlowStatus, err error) {
	fmt.Printf("Checking status of %s %s...\n", h.project.Name, h.flow.FlowID)
	client := h.client
	executions, err := client.FlowExecutions(h.project.Name, h.flow.FlowID)
	if err != nil {
		return status, err
	}

	if len(executions) == 0 {
		log.Printf("no executions")
		return status, nil
	}

	status.Health = executions.Health()
	histogram := executions.Histogram()
	details := executions.HistogramDetails(int(h.histogramCount))

	fmt.Printf("%-16s %s\n", "Job health:", status.Health.Colored())
	fmt.Printf("%-16s %d failures, %d successes, %d running, %d total\n", "Stats:", histogram.Failures, histogram.Successes, histogram.Running, histogram.Total)
	lastSuccessMessage := fmt.Sprintf("none in the last %d executions", len(executions))
	if histogram.LastSuccess != nil {
		lastSuccessMessage = humanize.Time(*histogram.LastSuccess)
	}
	fmt.Printf("%-16s %s\n", "Last success:", lastSuccessMessage)

	if err := h.printSchedule(); err != nil {
		return status, err
	}

	fmt.Printf("Histogram:       %s\n", histogram.Histogram)
	for _, l := range details {
		fmt.Printf("%-16s %s\n", " ", l)
	}

	if status.Health == azkaban.Critical {
		fmt.Println()

		status.LastExecution = executions.MostRecentExecution()
		executionID := executions.MostRecentExecution().ID

		flowExecStatus, err := client.FlowExecutionStatus(executionID)
		if err != nil {
			return status, err
		}

		for _, n := range flowExecStatus.Nodes {
			if n.Status.IsFailure() {
				status.FailedJob = n
				break
			}
		}
	}

	return status, nil
}

func printActionTerm() {
	fmt.Println("Actions: do nothing|status|restart|unschedule|logs")
	fmt.Print("> ")
}

type FlowStatus struct {
	Project       azkaban.Project
	Flow          azkaban.Flow
	Health        azkaban.Health
	LastExecution azkaban.Execution
	FailedJob     azkaban.JobStatus
}
