package azkaban

import (
	"fmt"

	"github.com/fatih/color"
)

type Status string

func (s Status) Colored() string {
	return s.ColorFunc()((fmt.Sprintf("%-9s", s)))
}

func (s Status) ColorFunc() func(string, ...interface{}) string {
	colorFunc := func(s string, x ...interface{}) string { return s }
	switch s {
	case "SUCCEEDED":
		colorFunc = color.GreenString
	case "FAILED":
		colorFunc = color.RedString
	case "RUNNING":
		colorFunc = color.CyanString
	case "CANCELLED":
		colorFunc = color.MagentaString
	case "PREPARING":
		colorFunc = color.YellowString
	default:
		colorFunc = color.WhiteString
	}

	return colorFunc
}

func (s Status) IsFailure() bool {
	return s == "FAILED"
}

func (s Status) IsSuccess() bool {
	return s == "SUCCEEDED"
}
