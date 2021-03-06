// Copyright (c) 2013 The meeko AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/meeko/meekod/supervisor/data"

	"github.com/tchap/gocli"
	"github.com/wsxiaoys/terminal/color"
)

func init() {
	app.MustRegisterSubcommand(&gocli.Command{
		UsageLine: "status [ALIAS]",
		Short:     "show agent status",
		Long: `
  When used without any argument, this command lists all installed agents
  together with their statuses. The status is one of the following:

    * stopped - the agent is stopped
    * running - the agent is running
    * crashed - the agent process returned a non-zero exit status
    * killed  - the agent process didn't exit cleanly and had to be killed

  When ALIS is defined, only the status of the chosen agent is printed.
        `,
		Action: runStatus,
	})
}

func runStatus(cmd *gocli.Command, args []string) {
	if len(args) > 1 {
		cmd.Usage()
		os.Exit(2)
	}

	if err := _runStatus(args); err != nil {
		os.Exit(1)
	}
}

func _runStatus(args []string) error {
	// Read the config file.
	cfg, err := LoadConfig(flagConfig)
	if err != nil {
		return err
	}

	// Send the status request to the server.
	statArgs := data.StatusArgs{
		Token: []byte(cfg.ManagementToken),
	}
	if len(args) == 1 {
		statArgs.Alias = args[0]
	}

	var reply data.StatusReply
	err = SendRequest(cfg.Address, cfg.AccessToken, MethodStatus, statArgs, &reply)
	if err != nil {
		return err
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}

	// Print the reply.
	if len(args) == 1 {
		fmt.Println(colorStatus(reply.Status))
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
	defer tw.Flush()

	fmt.Fprintln(tw, "")
	fmt.Fprintln(tw, "ALIAS\tSTATUS")
	fmt.Fprintln(tw, "=====\t======")

	for k, v := range reply.Statuses {
		fmt.Fprintf(tw, "%s\t%s\n", k, colorStatus(v))
	}

	fmt.Fprintln(tw, "")
	return nil
}

func colorStatus(status string) string {
	switch status {
	case "stopped":
		return "stopped"
	case "running":
		return color.Sprint("@{g}running@{|}")
	case "crashed":
		return color.Sprint("@{r}crashed@{|}")
	case "killed":
		return color.Sprint("@{m}killed@{|}")
	}

	panic("Unknown status returned")
}
