// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ChannelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Management of channels",
}

var ChannelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a channel",
	Long:  `Create a channel.`,
	Example: `  channel create --team myteam --name mynewchannel --display_name "My New Channel"
  channel create --team myteam --name mynewprivatechannel --display_name "My New Private Channel" --private`,
	RunE: withClient(createChannelCmdF),
}

// ChannelRenameCmd is used to change name and/or display name of an existing channel.
var ChannelRenameCmd = &cobra.Command{
	Use:   "rename [channel]",
	Short: "Rename channel",
	Long:  `Rename an existing channel.`,
	Example: `  channel rename myteam:oldchannel --name 'new-channel' --display_name 'New Display Name'
  channel rename myteam:oldchannel --name 'new-channel'
  channel rename myteam:oldchannel --display_name 'New Display Name'`,
	Args: cobra.ExactArgs(1),
	RunE: withClient(renameChannelCmdF),
}

var RemoveChannelUsersCmd = &cobra.Command{
	Use:   "remove [channel] [users]",
	Short: "Remove users from channel",
	Long:  "Remove some users from channel",
	Example: `  channel remove myteam:mychannel user@example.com username
  channel remove myteam:mychannel --all-users`,
	Deprecated: "please use \"users remove\" instead",
	RunE:       withClient(channelUsersRemoveCmdF),
}

var AddChannelUsersCmd = &cobra.Command{
	Use:        "add [channel] [users]",
	Short:      "Add users to channel",
	Long:       "Add some users to channel",
	Example:    "  channel add myteam:mychannel user@example.com username",
	Deprecated: "please use \"users add\" instead",
	RunE:       withClient(channelUsersAddCmdF),
}

var ArchiveChannelsCmd = &cobra.Command{
	Use:   "archive [channels]",
	Short: "Archive channels",
	Long: `Archive some channels.
Archive a channel along with all related information including posts from the database.
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel archive myteam:mychannel",
	RunE:    withClient(archiveChannelsCmdF),
}

var DeleteChannelsCmd = &cobra.Command{
	Use:   "delete [channels]",
	Short: "Delete channels",
	Long: `Permanently delete some channels.
Permanently deletes one or multiple channels along with all related information including posts from the database.`,
	Example: "  channel delete myteam:mychannel",
	Args:    cobra.MinimumNArgs(1),
	RunE:    withClient(deleteChannelsCmdF),
}

// ListChannelsCmd is a command which lists all the channels of team(s) in a server.
var ListChannelsCmd = &cobra.Command{
	Use:   "list [teams]",
	Short: "List all channels on specified teams.",
	Long: `List all channels on specified teams.
Archived channels are appended with ' (archived)'.
Private channels the user is a member of or has access to are appended with ' (private)'.`,
	Example: "  channel list myteam",
	Args:    cobra.MinimumNArgs(1),
	RunE:    withClient(listChannelsCmdF),
}

var ModifyChannelCmd = &cobra.Command{
	Use:   "modify [channel] [flags]",
	Short: "Modify a channel's public/private type",
	Long: `Change the Public/Private type of a channel.
Channel can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: `  channel modify myteam:mychannel --private
  channel modify channelId --public`,
	Args:   cobra.ExactArgs(1),
	PreRun: disableLocalPrecheck,
	RunE:   withClient(modifyChannelCmdF),
}

var RestoreChannelsCmd = &cobra.Command{
	Use:        "restore [channels]",
	Deprecated: "please use \"unarchive\" instead",
	Short:      "Restore some channels",
	Long: `Restore a previously deleted channel
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel restore myteam:mychannel",
	PreRun:  disableLocalPrecheck,
	RunE:    withClient(unarchiveChannelsCmdF),
}

var UnarchiveChannelCmd = &cobra.Command{
	Use:   "unarchive [channels]",
	Short: "Unarchive some channels",
	Long: `Unarchive a previously archived channel
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel unarchive myteam:mychannel",
	PreRun:  disableLocalPrecheck,
	RunE:    withClient(unarchiveChannelsCmdF),
}

var MakeChannelPrivateCmd = &cobra.Command{
	Use:   "make_private [channel]",
	Short: "Set a channel's type to private",
	Long: `Set the type of a channel from Public to Private.
Channel can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel make_private myteam:mychannel",
	PreRun:  disableLocalPrecheck,
	RunE:    withClient(makeChannelPrivateCmdF),
}

var SearchChannelCmd = &cobra.Command{
	Use:   "search [channel]\n  mmctl search --team [team] [channel]",
	Short: "Search a channel",
	Long: `Search a channel by channel name.
Channel can be specified by team. ie. --team myteam mychannel or by team ID.`,
	Example: `  channel search mychannel
  channel search --team myteam mychannel`,
	Args: cobra.ExactArgs(1),
	RunE: withClient(searchChannelCmdF),
}

var MoveChannelCmd = &cobra.Command{
	Use:   "move [team] [channels]",
	Short: "Moves channels to the specified team",
	Long: `Moves the provided channels to the specified team.
Validates that all users in the channel belong to the target team. Incoming/Outgoing webhooks are moved along with the channel.
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel move newteam oldteam:mychannel",
	Args:    cobra.MinimumNArgs(2),
	RunE:    withClient(moveChannelCmdF),
}

func init() {
	ChannelCreateCmd.Flags().String("name", "", "Channel Name")
	ChannelCreateCmd.Flags().String("display_name", "", "Channel Display Name")
	ChannelCreateCmd.Flags().String("team", "", "Team name or ID")
	ChannelCreateCmd.Flags().String("header", "", "Channel header")
	ChannelCreateCmd.Flags().String("purpose", "", "Channel purpose")
	ChannelCreateCmd.Flags().Bool("private", false, "Create a private channel.")

	ModifyChannelCmd.Flags().Bool("private", false, "Convert the channel to a private channel")
	ModifyChannelCmd.Flags().Bool("public", false, "Convert the channel to a public channel")

	ChannelRenameCmd.Flags().String("name", "", "Channel Name")
	ChannelRenameCmd.Flags().String("display_name", "", "Channel Display Name")

	RemoveChannelUsersCmd.Flags().Bool("all-users", false, "Remove all users from the indicated channel.")

	SearchChannelCmd.Flags().String("team", "", "Team name or ID")

	MoveChannelCmd.Flags().Bool("force", false, "Remove users that are not members of target team before moving the channel.")

	DeleteChannelsCmd.Flags().Bool("confirm", false, "Confirm you really want to delete the channel and a DB backup has been performed.")

	ChannelCmd.AddCommand(
		ChannelCreateCmd,
		RemoveChannelUsersCmd,
		AddChannelUsersCmd,
		ArchiveChannelsCmd,
		ListChannelsCmd,
		RestoreChannelsCmd,
		UnarchiveChannelCmd,
		MakeChannelPrivateCmd,
		ModifyChannelCmd,
		ChannelRenameCmd,
		SearchChannelCmd,
		MoveChannelCmd,
		DeleteChannelsCmd,
	)

	RootCmd.AddCommand(ChannelCmd)
}

func createChannelCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

	name, errn := cmd.Flags().GetString("name")
	if errn != nil || name == "" {
		return errors.New("name is required")
	}
	displayname, errdn := cmd.Flags().GetString("display_name")
	if errdn != nil || displayname == "" {
		return errors.New("display Name is required")
	}
	teamArg, errteam := cmd.Flags().GetString("team")
	if errteam != nil || teamArg == "" {
		return errors.New("team is required")
	}
	header, _ := cmd.Flags().GetString("header")
	purpose, _ := cmd.Flags().GetString("purpose")
	useprivate, _ := cmd.Flags().GetBool("private")

	channelType := model.CHANNEL_OPEN
	if useprivate {
		channelType = model.CHANNEL_PRIVATE
	}

	team := getTeamFromTeamArg(c, teamArg)
	if team == nil {
		return errors.Errorf("unable to find team: %s", teamArg)
	}

	channel := &model.Channel{
		TeamId:      team.Id,
		Name:        name,
		DisplayName: displayname,
		Header:      header,
		Purpose:     purpose,
		Type:        channelType,
		CreatorId:   "",
	}

	newChannel, response := c.CreateChannel(channel)
	if response.Error != nil {
		return response.Error
	}

	printer.PrintT("New channel {{.Name}} successfully created", newChannel)

	return nil
}

func archiveChannelsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("enter at least one channel to archive")
	}

	channels := getChannelsFromChannelArgs(c, args)
	for i, channel := range channels {
		if channel == nil {
			printer.PrintError("Unable to find channel '" + args[i] + "'")
			continue
		}
		if _, response := c.DeleteChannel(channel.Id); response.Error != nil {
			printer.PrintError("Unable to archive channel '" + channel.Name + "' error: " + response.Error.Error())
		}
	}

	return nil
}

func listChannelsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	teams := getTeamsFromTeamArgs(c, args)
	for i, team := range teams {
		if team == nil {
			printer.PrintError("Unable to find team '" + args[i] + "'")
			continue
		}

		publicChannels, response := c.GetPublicChannelsForTeam(team.Id, 0, 10000, "")
		if response.Error != nil {
			printer.PrintError("Unable to list public channels for '" + args[i] + "'. Error: " + response.Error.Error())
		}
		for _, channel := range publicChannels {
			printer.PrintT("{{.Name}}", channel)
		}

		deletedChannels, response := c.GetDeletedChannelsForTeam(team.Id, 0, 10000, "")
		if response.Error != nil {
			printer.PrintError("Unable to list archived channels for '" + args[i] + "'. Error: " + response.Error.Error())
		}
		for _, channel := range deletedChannels {
			printer.PrintT("{{.Name}} (archived)", channel)
		}

		privateChannels, err := getPrivateChannels(c, team.Id)
		if err != nil {
			printer.PrintError("Unable to list private channels for '" + args[i] + "'. Error: " + err.Error())
		}
		for _, channel := range privateChannels {
			printer.PrintT("{{.Name}} (private)", channel)
		}
	}

	return nil
}

func unarchiveChannelsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("enter at least one channel")
	}

	channels := getChannelsFromChannelArgs(c, args)
	for i, channel := range channels {
		if channel == nil {
			printer.PrintError("Unable to find channel '" + args[i] + "'")
			continue
		}
		if _, response := c.RestoreChannel(channel.Id); response.Error != nil {
			printer.PrintError("Unable to unarchive channel '" + args[i] + "'. Error: " + response.Error.Error())
		}
	}

	return nil
}

func makeChannelPrivateCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("enter one channel to modify")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.Errorf("unable to find channel %q", args[0])
	}

	if !(channel.Type == model.CHANNEL_OPEN) {
		return errors.New("you can only change the type of public channels")
	}

	if _, response := c.UpdateChannelPrivacy(channel.Id, model.CHANNEL_PRIVATE); response.Error != nil {
		return response.Error
	}

	return nil
}

func modifyChannelCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	public, _ := cmd.Flags().GetBool("public")
	private, _ := cmd.Flags().GetBool("private")

	if public == private {
		return errors.New("you must specify only one of --public or --private")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.Errorf("unable to find channel %q", args[0])
	}

	if !(channel.Type == model.CHANNEL_OPEN || channel.Type == model.CHANNEL_PRIVATE) {
		return errors.New("you can only change the type of public/private channels")
	}

	privacy := model.CHANNEL_OPEN
	if private {
		privacy = model.CHANNEL_PRIVATE
	}

	if _, response := c.UpdateChannelPrivacy(channel.Id, privacy); response.Error != nil {
		return errors.Errorf("failed to update channel (%q) privacy: %s", args[0], response.Error.Error())
	}

	return nil
}

func renameChannelCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	existingTeamChannel := args[0]

	newChannelName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	newDisplayName, err := cmd.Flags().GetString("display_name")
	if err != nil {
		return err
	}

	// At least one of display name or name flag must be present
	if newDisplayName == "" && newChannelName == "" {
		return errors.New("require at least one flag to rename channel, either 'name' or 'display_name'")
	}

	channel := getChannelFromChannelArg(c, existingTeamChannel)
	if channel == nil {
		return errors.Errorf("unable to find channel from %q", existingTeamChannel)
	}

	channelPatch := &model.ChannelPatch{}
	if newChannelName != "" {
		channelPatch.Name = &newChannelName
	}
	if newDisplayName != "" {
		channelPatch.DisplayName = &newDisplayName
	}

	// Using PatchChannel API to rename channel
	updatedChannel, response := c.PatchChannel(channel.Id, channelPatch)
	if response.Error != nil {
		return errors.Errorf("cannot rename channel %q, error: %s", channel.Name, response.Error.Error())
	}

	printer.PrintT("'{{.Name}}' channel renamed", updatedChannel)
	return nil
}

func searchChannelCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

	var channel *model.Channel

	if teamArg, _ := cmd.Flags().GetString("team"); teamArg != "" {
		team := getTeamFromTeamArg(c, teamArg)
		if team == nil {
			return errors.Errorf("team %s was not found", teamArg)
		}

		var response *model.Response
		channel, response = c.GetChannelByName(args[0], team.Id, "")
		if response.Error != nil {
			return response.Error
		}
		if channel == nil {
			return errors.Errorf("channel %s was not found in team %s", args[0], teamArg)
		}
	} else {
		teams, response := c.GetAllTeams("", 0, 9999)
		if response.Error != nil {
			return response.Error
		}

		for _, team := range teams {
			channel, _ = c.GetChannelByName(args[0], team.Id, "")
			if channel != nil && channel.Name == args[0] {
				break
			}
		}

		if channel == nil {
			return errors.Errorf("channel %q was not found in any team", args[0])
		}
	}

	if channel.DeleteAt > 0 {
		printer.PrintT("Channel Name :{{.Name}}, Display Name :{{.DisplayName}}, Channel ID :{{.Id}} (archived)", channel)
	} else {
		printer.PrintT("Channel Name :{{.Name}}, Display Name :{{.DisplayName}}, Channel ID :{{.Id}}", channel)
	}
	return nil
}

func moveChannelCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	team := getTeamFromTeamArg(c, args[0])
	if team == nil {
		return fmt.Errorf("unable to find destination team %q", args[0])
	}

	channels := getChannelsFromChannelArgs(c, args[1:])
	for i, channel := range channels {
		if channel == nil {
			printer.PrintError(fmt.Sprintf("Unable to find channel %q", args[i+1]))
			continue
		}

		if channel.TeamId == team.Id {
			continue
		}

		newChannel, resp := c.MoveChannel(channel.Id, team.Id, force)
		if resp.Error != nil {
			printer.PrintError(fmt.Sprintf("unable to move channel %q: %s", channel.Name, resp.Error))
			continue
		}
		printer.PrintT(fmt.Sprintf("Moved channel {{.Name}} to %q ({{.TeamId}}) from %s.", team.Name, channel.TeamId), newChannel)
	}
	return nil
}

func getPrivateChannels(c client.Client, teamID string) ([]*model.Channel, *model.AppError) {
	allPrivateChannels, response := c.GetPrivateChannelsForTeam(teamID, 0, 10000, "")
	// local mode admin result is a superset of user result
	// if see an error and we're in local mode we can return early
	if response.Error == nil || viper.GetBool("local") {
		return allPrivateChannels, response.Error
	}

	// We are definitely not in local mode here so we can safely use "GetChannelsForTeamForUser"
	// and "me" for userId
	allChannels, response := c.GetChannelsForTeamForUser(teamID, "me", false, "")
	if response.Error != nil {
		if response.StatusCode == http.StatusNotFound { // user doesn't belong to any channels
			return nil, nil
		}
		return nil, response.Error
	}
	privateChannels := make([]*model.Channel, 0, len(allChannels))
	for _, channel := range allChannels {
		if channel.Type != model.CHANNEL_PRIVATE {
			continue
		}
		privateChannels = append(privateChannels, channel)
	}
	return privateChannels, nil
}

func getChannelDeleteConfirmation() error {
	var confirm string
	fmt.Println("Have you performed a database backup? (YES/NO): ")
	fmt.Scanln(&confirm)

	if confirm != "YES" {
		return errors.New("aborted: You did not answer YES exactly, in all capitals")
	}
	fmt.Println("Are you sure you want to delete the channels specified? All data will be permanently deleted? (YES/NO): ")
	fmt.Scanln(&confirm)
	if confirm != "YES" {
		return errors.New("aborted: You did not answer YES exactly, in all capitals")
	}
	return nil
}

func deleteChannelsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	confirmFlag, _ := cmd.Flags().GetBool("confirm")
	if !confirmFlag {
		if err := getChannelDeleteConfirmation(); err != nil {
			return err
		}
	}

	channels := getChannelsFromChannelArgs(c, args)
	for i, channel := range channels {
		if channel == nil {
			printer.PrintError("Unable to find channel '" + args[i] + "'")
			continue
		}
		if _, response := c.PermanentDeleteChannel(channel.Id); response.Error != nil {
			printer.PrintError("Unable to delete channel '" + channel.Name + "' error: " + response.Error.Error())
		} else {
			printer.PrintT("Deleted channel '{{.Name}}'", channel)
		}
	}
	return nil
}
