package azkaban

import (
	"testing"
	"time"
)

func TestAzkabanStringTimeUnmarshal(t *testing.T) {
	input := "2017-04-01 09:00:00"

	a := &AzkabanStringTime{}
	a.UnmarshalJSON([]byte(input))
	println(a.Time().Format(time.RFC3339))

}
