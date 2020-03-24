// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var BotCmd = &cobra.Command{
	Use:   "bot",
	Short: "Management of bots",
}

var CreateBotCmd = &cobra.Command{
	Use:     "create [username]",
	Short:   "Create bot",
	Long:    "Create bot.",
	Example: `  bot create testbot`,
	RunE:    withClient(botCreateCmdF),
	Args:    cobra.ExactArgs(1),
}

var UpdateBotCmd = &cobra.Command{
	Use:     "update [username]",
	Short:   "Update bot",
	Long:    "Update bot information.",
	Example: `  bot update testbot --username newbotusername`,
	RunE:    withClient(botUpdateCmdF),
	Args:    cobra.ExactArgs(1),
}

var ListBotCmd = &cobra.Command{
	Use:     "list",
	Short:   "List bots",
	Long:    "List the bots users.",
	Example: `  bot list`,
	RunE:    withClient(botListCmdF),
	Args:    cobra.ExactArgs(0),
}

var DisableBotCmd = &cobra.Command{
	Use:     "disable [username]",
	Short:   "Disable bot",
	Long:    "Disable a disabled bot",
	Example: `  bot enable testbot`,
	RunE:    withClient(botDisableCmdF),
	Args:    cobra.MinimumNArgs(1),
}

var EnableBotCmd = &cobra.Command{
	Use:     "enable [username]",
	Short:   "Enable bot",
	Long:    "Enable a disabled bot",
	Example: `  bot enable testbot`,
	RunE:    withClient(botEnableCmdF),
	Args:    cobra.MinimumNArgs(1),
}

var AssignBotCmd = &cobra.Command{
	Use:     "assign [bot-username] [new-owner-username]",
	Short:   "Assign bot",
	Long:    "Assign the ownership of a bot to another user",
	Example: `  bot assign testbot user2`,
	RunE:    withClient(botAssignCmdF),
	Args:    cobra.ExactArgs(2),
}

func init() {
	CreateBotCmd.Flags().String("display-name", "", "Optional. The display name for the new bot.")
	CreateBotCmd.Flags().String("description", "", "Optional. The description text for the new bot.")
	ListBotCmd.Flags().Bool("orphaned", false, "Optional. Only show orphaned bots.")
	ListBotCmd.Flags().Bool("all", false, "Optional. Show all bots (including deleleted and orphaned).")
	UpdateBotCmd.Flags().String("username", "", "Optional. The new username for the bot.")
	UpdateBotCmd.Flags().String("display-name", "", "Optional. The new display name for the bot.")
	UpdateBotCmd.Flags().String("description", "", "Optional. The new description text for the bot.")

	BotCmd.AddCommand(
		CreateBotCmd,
		UpdateBotCmd,
		ListBotCmd,
		EnableBotCmd,
		DisableBotCmd,
		AssignBotCmd,
	)

	RootCmd.AddCommand(BotCmd)
}

func botCreateCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	username := args[0]
	displayName, _ := cmd.Flags().GetString("display-name")
	description, _ := cmd.Flags().GetString("description")

	bot, res := c.CreateBot(&model.Bot{
		Username:    username,
		DisplayName: displayName,
		Description: description,
	})
	if err := res.Error; err != nil {
		return errors.Errorf("could not create bot: %s", err)
	}

	printer.PrintT("Created bot {{.UserId}}", bot)

	return nil
}

func botUpdateCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	user := getUserFromUserArg(c, args[0])
	if user == nil {
		return errors.New("unable to find user '" + args[0] + "'")
	}
	patch := model.BotPatch{}
	username, err := cmd.Flags().GetString("username")
	if err == nil && cmd.Flags().Changed("username") {
		patch.Username = &username
	}
	displayName, err := cmd.Flags().GetString("display-name")
	if err == nil && cmd.Flags().Changed("display-name") {
		patch.DisplayName = &displayName
	}
	description, err := cmd.Flags().GetString("description")
	if err == nil && cmd.Flags().Changed("description") {
		patch.Description = &description
	}

	bot, res := c.PatchBot(user.Id, &patch)
	if err := res.Error; err != nil {
		return errors.Errorf("could not update bot: %s", err)
	}

	printer.PrintT("Updated bot {{.UserId}} ({{.Username}})", bot)

	return nil
}

func botListCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	orphaned, _ := cmd.Flags().GetBool("orphaned")
	all, _ := cmd.Flags().GetBool("all")

	page := 0
	perPage := 200
	tpl := `{{.UserId}}: {{.Username}}`
	for {
		var bots []*model.Bot
		var res *model.Response
		if all { //nolint:ifElseChain
			bots, res = c.GetBotsIncludeDeleted(page, perPage, "")
		} else if orphaned {
			bots, res = c.GetBotsOrphaned(page, perPage, "")
		} else {
			bots, res = c.GetBots(page, perPage, "")
		}
		if res.Error != nil {
			return errors.Wrap(res.Error, "Failed to fetch bots")
		}

		if len(bots) == 0 {
			break
		}

		userIds := []string{}
		for _, bot := range bots {
			userIds = append(userIds, bot.OwnerId)
		}

		users, res := c.GetUsersByIds(userIds)
		if res.Error != nil {
			return errors.Wrap(res.Error, "Failed to fetch bots")
		}

		usersByID := map[string]*model.User{}
		for _, user := range users {
			usersByID[user.Id] = user
		}

		for _, bot := range bots {
			owner := usersByID[bot.OwnerId]
			tplExtraText := fmt.Sprintf("(Owner by %s, {{if ne .DeleteAt 0}}Disabled{{else}}Enabled{{end}}{{if ne %d 0}}, Orphaned{{end}})", owner.Username, owner.DeleteAt)
			printer.PrintT(tpl+tplExtraText, bot)
		}

		page++
	}

	return nil
}

func botEnableCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	users := getUsersFromUserArgs(c, args)
	for i, user := range users {
		if user == nil {
			printer.PrintError(fmt.Sprintf("can't find user '%v'", args[i]))
			continue
		}

		bot, res := c.EnableBot(user.Id)
		if err := res.Error; err != nil {
			printer.PrintError(fmt.Sprintf("could not enable bot '%v'", args[i]))
			continue
		}

		printer.PrintT("Enabled bot {{.UserId}} ({{.Username}})", bot)
	}

	return nil
}

func botDisableCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	users := getUsersFromUserArgs(c, args)
	for i, user := range users {
		if user == nil {
			printer.PrintError(fmt.Sprintf("can't find user '%v'", args[i]))
			continue
		}

		bot, res := c.DisableBot(user.Id)
		if err := res.Error; err != nil {
			printer.PrintError(fmt.Sprintf("could not disable bot '%v'", args[i]))
			continue
		}

		printer.PrintT("Disabled bot {{.UserId}} ({{.Username}})", bot)
	}

	return nil
}

func botAssignCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	botUser := getUserFromUserArg(c, args[0])
	if botUser == nil {
		return errors.New("unable to find user '" + args[0] + "'")
	}
	newOwnerUser := getUserFromUserArg(c, args[1])
	if newOwnerUser == nil {
		return errors.New("unable to find user '" + args[1] + "'")
	}

	newBot, res := c.AssignBot(botUser.Id, newOwnerUser.Id)
	if err := res.Error; err != nil {
		return errors.Errorf("can not assign bot '%s' to user '%s'", args[0], args[1])
	}

	printer.PrintT("The bot {{.UserId}} ({{.Username}}) now belongs to the user "+newOwnerUser.Username, newBot)
	return nil
}
