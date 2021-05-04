// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

var RolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Manage user roles",
}

var RolesSystemAdminCmd = &cobra.Command{
	Use:   "system_admin [users]",
	Short: "Set a user as system admin",
	Long:  "Make some users system admins.",
	Example: `  # You can make one user a sysadmin
  $ mmctl roles system_admin john_doe

  # Or promote multiple users at the same time
  $ mmctl roles system_admin john_doe jane_doe`,
	RunE: withClient(rolesSystemAdminCmdF),
	Args: cobra.MinimumNArgs(1),
}

var RolesMemberCmd = &cobra.Command{
	Use:   "member [users]",
	Short: "Remove system admin privileges",
	Long:  "Remove system admin privileges from some users.",
	Example: `  # You can remove admin privileges from one user
  $ mmctl roles member john_doe

  # Or demote multiple users at the same time
  $ mmctl roles member john_doe jane_doe`,
	RunE: withClient(rolesMemberCmdF),
	Args: cobra.MinimumNArgs(1),
}

func init() {
	RolesCmd.AddCommand(
		RolesSystemAdminCmd,
		RolesMemberCmd,
	)

	RootCmd.AddCommand(RolesCmd)
}

func rolesSystemAdminCmdF(c client.Client, _ *cobra.Command, args []string) error {
	users := getUsersFromUserArgs(c, args)
	for i, user := range users {
		if user == nil {
			printer.PrintError(fmt.Sprintf("unable to find user %q", args[i]))
			continue
		}

		systemAdmin := false
		roles := strings.Fields(user.Roles)
		for _, role := range roles {
			if role == model.SYSTEM_ADMIN_ROLE_ID {
				systemAdmin = true
			}
		}

		if !systemAdmin {
			roles = append(roles, model.SYSTEM_ADMIN_ROLE_ID)
			if _, resp := c.UpdateUserRoles(user.Id, strings.Join(roles, " ")); resp.Error != nil {
				printer.PrintError(fmt.Sprintf("can't update roles for user %q: %s", args[i], resp.Error))
				continue
			}

			printer.Print(fmt.Sprintf("System admin role assigned to user %q", args[i]))
		}
	}

	return nil
}

func rolesMemberCmdF(c client.Client, _ *cobra.Command, args []string) error {
	users := getUsersFromUserArgs(c, args)
	for i, user := range users {
		if user == nil {
			printer.PrintError(fmt.Sprintf("unable to find user %q", args[i]))
			continue
		}

		shouldRemoveSysadmin := false
		var newRoles []string

		roles := strings.Fields(user.Roles)
		for _, role := range roles {
			switch role {
			case model.SYSTEM_ADMIN_ROLE_ID:
				shouldRemoveSysadmin = true
			default:
				newRoles = append(newRoles, role)
			}
		}

		if shouldRemoveSysadmin {
			if _, resp := c.UpdateUserRoles(user.Id, strings.Join(newRoles, " ")); resp.Error != nil {
				printer.PrintError(fmt.Sprintf("can't update roles for user %q: %s", args[i], resp.Error))
				continue
			}

			printer.Print(fmt.Sprintf("System admin role revoked for user %q", args[i]))
		}
	}

	return nil
}
