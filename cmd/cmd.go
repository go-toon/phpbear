package cmd

import (
	"github.com/urfave/cli"
)

var (
	FrameworkGit  string
	FrameworkTag  string
	ProjectGit    string
	ProjectName   string
	ProjectAlias  string
	ProjectTag    string
	FeatureBranch string
)

type PhpBearOptions struct {
	FrameworkGit string `json:"framework_git"`
	FrameworkTag string `json:"framework_tag"`
	ProjectGit   string `json:"project_git"`
	InitTime     int64  `json:"init_time,omitempty"`
}

func stringFlag(name, value, usage string) cli.StringFlag {
	return cli.StringFlag{
		Name:  name,
		Value: value,
		Usage: usage,
	}
}
