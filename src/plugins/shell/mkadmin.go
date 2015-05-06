package shell

/**

Provides the ability to list users

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"fmt"
)

// The clear command
type GrantAdmin struct {
}

func (h GrantAdmin) Help() string {
	return "Grants admin to a user: 'grantadmin username'"
}

func (h GrantAdmin) Usage() string {
	return ""
}

func (h GrantAdmin) Execute(shell *Shell, args []string) {
	if len(args) < 2 {
		fmt.Println(Red + "Must supply a name" + Reset)
		return
	}

	operator := shell.operator

	err := operator.SetAdmin(args[1], true)
	if shell.PrintError(err) {
		return
	}

	fmt.Println(Green + "Granted admin to: " + args[1] + Reset)
}

func (h GrantAdmin) Name() string {
	return "mkadmin"
}
