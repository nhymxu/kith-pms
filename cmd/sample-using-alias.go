package main

import (
	"context"

	"github.com/urfave/cli/v3"

	"rootPrj/app/sample-using-alias"
)

func sampleAliasCommand() *cli.Command {
	return &cli.Command{
		Name:        "sample-alias",
		Usage:       "Sample using alias",
		Description: `Sample using alias, without long project name`,
		Action: func(_ context.Context, _ *cli.Command) error {
			sample_using_alias.Run()
			return nil
		},
	}
}
