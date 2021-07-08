// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"strings"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var ChannelUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Management of channel users",
}

var ChannelUsersAddCmd = &cobra.Command{
	Use:     "add [channel] [users]",
	Short:   "Add users to channel",
	Long:    "Add some users to channel",
	Example: "  channel users add myteam:mychannel user@example.com username",
	RunE:    withClient(channelUsersAddCmdF),
}

var ChannelUsersListCmd = &cobra.Command{
	Use:     "list [channel] [users]",
	Short:   "List users with access to a private channel",
	Long:    "List users with access to a private channel",
	Example: "  channel users list myteam:mychannel",
	RunE:    withClient(channelUsersListCmdF),
}

var ChannelUsersRemoveCmd = &cobra.Command{
	Use:   "remove [channel] [users]",
	Short: "Remove users from channel",
	Long:  "Remove some users from channel",
	Example: `  channel users remove myteam:mychannel user@example.com username
  channel users remove myteam:mychannel --all-users`,
	RunE: withClient(channelUsersRemoveCmdF),
}

func init() {
	ChannelUsersRemoveCmd.Flags().Bool("all-users", false, "Remove all users from the indicated channel.")
	ChannelUsersListCmd.Flags().String("display", "email", "Display the 'username' or 'email' or 'id' of the channel members.")

	ChannelUsersCmd.AddCommand(
		ChannelUsersListCmd,
		ChannelUsersAddCmd,
		ChannelUsersRemoveCmd,
	)

	ChannelCmd.AddCommand(ChannelUsersCmd)
}

func channelUsersAddCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("not enough arguments")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.Errorf("unable to find channel %q", args[0])
	}

	users := getUsersFromUserArgs(c, args[1:])
	for i, user := range users {
		addUserToChannel(c, channel, user, args[i+1])
	}

	return nil
}

func addUserToChannel(c client.Client, channel *model.Channel, user *model.User, userArg string) {
	if user == nil {
		printer.PrintError("Can't find user '" + userArg + "'")
		return
	}
	if _, response := c.AddChannelMember(channel.Id, user.Id); response.Error != nil {
		printer.PrintError("Unable to add '" + userArg + "' to " + channel.Name + ". Error: " + response.Error.Error())
	}
}

func channelUsersRemoveCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	allUsers, _ := cmd.Flags().GetBool("all-users")

	if allUsers && len(args) != 1 {
		return errors.New("individual users must not be specified in conjunction with the --all-users flag")
	}

	if !allUsers && len(args) < 2 {
		return errors.New("you must specify some users to remove from the channel, or use the --all-users flag to remove them all")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.Errorf("unable to find channel %q", args[0])
	}

	if allUsers {
		removeAllUsersFromChannel(c, channel)
	} else {
		for i, user := range getUsersFromUserArgs(c, args[1:]) {
			removeUserFromChannel(c, channel, user, args[i+1])
		}
	}

	return nil
}

func removeUserFromChannel(c client.Client, channel *model.Channel, user *model.User, userArg string) {
	if user == nil {
		printer.PrintError("Can't find user '" + userArg + "'")
		return
	}
	if _, response := c.RemoveUserFromChannel(channel.Id, user.Id); response.Error != nil {
		printer.PrintError("Unable to remove '" + userArg + "' from " + channel.Name + ". Error: " + response.Error.Error())
	}
}

func removeAllUsersFromChannel(c client.Client, channel *model.Channel) {
	members, response := c.GetChannelMembers(channel.Id, 0, 10000, "")
	if response.Error != nil {
		printer.PrintError("Unable to remove all users from " + channel.Name + ". Error: " + response.Error.Error())
	}

	for _, member := range *members {
		if _, response := c.RemoveUserFromChannel(channel.Id, member.UserId); response.Error != nil {
			printer.PrintError("Unable to remove '" + member.UserId + "' from " + channel.Name + ". Error: " + response.Error.Error())
		}
	}
}

func userIdsOfChannelMembers(c client.Client, channel *model.Channel) ([]string, error) {
	members, response := c.GetChannelMembers(channel.Id, 0, 10000, "")
	if response.Error != nil {
		printer.PrintError("Unable to list the users of " + channel.Name + ". Error: " + response.Error.Error())
		return nil, response.Error
	}

	var userIds []string
	for _, member := range *members {
		userIds = append(userIds, member.UserId)
	}
	return userIds, nil
}

func channelUsersListCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.Errorf("unable to find channel %q", args[0])
	}

	userIds, err := userIdsOfChannelMembers(c, channel)
	if err != nil {
		printer.PrintError("Unable to list the users of " + channel.Name + ". Error: " + err.Error())
		return err
	}

	display, _ := cmd.Flags().GetString("display")
	if display == "" {
		display = "email"
	}

	users, response := c.GetUsersByIds(userIds)
	if response.Error != nil {
		printer.PrintError("Unable to resolve the users of " + channel.Name + ". Error: " + response.Error.Error())
		return response.Error
	}

	var displayed []string
	if display == "username" {
		for _, user := range users {
			if user.DeleteAt == 0 {
				displayed = append(displayed, user.Username)
			}
		}
	} else if display == "id" {
		for _, user := range users {
			if user.DeleteAt == 0 {
				displayed = append(displayed, user.Id)
			}
		}
	} else {
		for _, user := range users {
			if user.DeleteAt == 0 {
				displayed = append(displayed, user.Email)
			}
		}
	}

	printer.Print(strings.Join(displayed, " "))
	return nil
}
