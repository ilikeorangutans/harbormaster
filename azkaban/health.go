package azkaban

import "github.com/fatih/color"

type Health string

func (h Health) Colored() string {
	switch h {
	case Healthy:
		return color.GreenString(string(h))
	case Concerning:
		return color.YellowString(string(h))
	case Critical:
		return color.RedString(string(h))
	default:
		return string(h)

	}
}

func (h Health) IsHealthy() bool {
	switch h {
	case Healthy:
		return true
	default:
		return false
	}
}

const (
	Healthy    Health = "healthy"
	Concerning        = "concerning"
	Critical          = "critical"
)
