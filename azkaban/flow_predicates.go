package azkaban

import (
	"log"
	"regexp"
	"strings"
)

type FlowPredicate func(Flow) bool

func MatchesAnyFlow() FlowPredicate {
	return func(_ Flow) bool { return true }
}
func MatchesFlowName(input string) FlowPredicate {
	if len(input) > 0 {
		regexString := input
		if strings.HasPrefix(regexString, "/") && strings.HasSuffix(regexString, "/") {
			regex, err := regexp.Compile(regexString[1 : len(input)-1])
			if err != nil {
				log.Fatal(err)
			}
			return func(f Flow) bool { return regex.MatchString(f.FlowID) }
		} else {
			return func(f Flow) bool { return strings.HasPrefix(f.FlowID, input) }
		}
	} else {
		return func(f Flow) bool { return true }
	}

}

// MatchesAll returns a FlowPredicate that returns true if a flow matches all the provided predicates.
func MatchesAll(predicates ...FlowPredicate) FlowPredicate {
	return func(f Flow) bool {
		for _, p := range predicates {
			if !p(f) {
				return false
			}
		}
		return true
	}
}
