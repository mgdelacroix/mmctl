// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestCreateIncomingWebhookCmd() {
	s.SetupTestHelper().InitBasic()

	oldEnablePostUsernameOverride := *s.th.App.Config().ServiceSettings.EnablePostUsernameOverride
	s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostUsernameOverride = true })
	defer s.th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.ServiceSettings.EnablePostUsernameOverride = oldEnablePostUsernameOverride
	})

	s.Run("Unprivileged user should not be able to create an incoming webhook", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("channel", s.th.BasicChannel.Id, "")
		cmd.Flags().String("user", s.th.BasicUser.Username, "")

		err := createIncomingWebhookCmdF(s.th.Client, cmd, []string{})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "appropriate permissions")
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to create webhook", printer.GetErrorLines()[0])
	})

	s.Run("Sysadmin should be able to create an incoming webhook without specifying an owner", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("channel", s.th.BasicChannel.Id, "")
		cmd.Flags().String("user", s.th.BasicUser.Username, "")

		err := createIncomingWebhookCmdF(s.th.SystemAdminClient, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		webhook := printer.GetLines()[0].(*model.IncomingWebhook)
		s.Require().Equal(s.th.BasicChannel.Id, webhook.ChannelId)
		s.Require().Equal(s.th.SystemAdminUser.Id, webhook.UserId)
		s.Require().Equal(s.th.BasicUser.Username, webhook.Username)

		_, appErr := s.th.App.GetIncomingWebhook(webhook.Id)
		s.Require().Nil(appErr)
	})

	s.Run("Local mode can't create an incoming webhook without specifying an owner", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("channel", s.th.BasicChannel.Id, "")
		cmd.Flags().String("user", s.th.BasicUser.Username, "")
		viper.Set("local", true)
		defer viper.Set("local", nil)

		err := createIncomingWebhookCmdF(s.th.LocalClient, cmd, []string{})
		s.Require().Error(err)
		s.Require().EqualError(err, "owner should be specified to run this command in local mode")
	})

	s.RunForSystemAdminAndLocal("Create incoming webhook from a specific user", func(c client.Client) {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("channel", s.th.BasicChannel.Id, "")
		cmd.Flags().String("user", s.th.BasicUser.Username, "")
		cmd.Flags().String("owner", s.th.BasicUser.Username, "")

		err := createIncomingWebhookCmdF(c, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		webhook := printer.GetLines()[0].(*model.IncomingWebhook)
		s.Require().Equal(s.th.BasicChannel.Id, webhook.ChannelId)
		s.Require().Equal(s.th.BasicUser.Id, webhook.UserId)
		s.Require().Equal(s.th.BasicUser.Username, webhook.Username)

		_, appErr := s.th.App.GetIncomingWebhook(webhook.Id)
		s.Require().Nil(appErr)
	})
}

func (s *MmctlE2ETestSuite) TestCreateOutgoingWebhookCmd() {
	s.SetupTestHelper().InitBasic()

	s.Run("Unprivileged user should not be able to create an outgoing webhook", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("team", s.th.BasicTeam.Id, "")
		cmd.Flags().String("user", s.th.BasicUser.Username, "")
		cmd.Flags().String("display-name", model.NewId(), "")
		cmd.Flags().String("channel", s.th.BasicChannel.Id, "")
		cmd.Flags().String("trigger-when", "exact", "")
		cmd.Flags().StringArray("url", []string{"http://example.com"}, "")

		err := createOutgoingWebhookCmdF(s.th.Client, cmd, []string{})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "appropriate permissions")
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to create outgoing webhook", printer.GetErrorLines()[0])
	})

	s.Run("Sysadmin should be able to create an outgoing webhook without specifying an owner", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("team", s.th.BasicTeam.Id, "")
		cmd.Flags().String("user", s.th.BasicUser.Username, "")
		cmd.Flags().String("display-name", model.NewId(), "")
		cmd.Flags().String("channel", s.th.BasicChannel.Id, "")
		cmd.Flags().String("trigger-when", "exact", "")
		cmd.Flags().StringArray("url", []string{"http://example.com"}, "")

		err := createOutgoingWebhookCmdF(s.th.SystemAdminClient, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		webhook := printer.GetLines()[0].(*model.OutgoingWebhook)
		s.Require().Equal(s.th.BasicTeam.Id, webhook.TeamId)
		s.Require().Equal(s.th.SystemAdminUser.Id, webhook.CreatorId)
		s.Require().Equal(s.th.BasicUser.Username, webhook.Username)
	})

	s.Run("Local mode can't create an outgoing webhook without specifying an owner", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("team", s.th.BasicTeam.Id, "")
		cmd.Flags().String("user", s.th.BasicUser.Username, "")
		cmd.Flags().String("display-name", model.NewId(), "")
		cmd.Flags().String("channel", s.th.BasicChannel.Id, "")
		cmd.Flags().String("trigger-when", "exact", "")
		cmd.Flags().StringArray("url", []string{"http://example.com"}, "")
		viper.Set("local", true)
		defer viper.Set("local", nil)

		err := createOutgoingWebhookCmdF(s.th.SystemAdminClient, cmd, []string{})
		s.Require().Error(err)
		s.Require().EqualError(err, "owner should be specified to run this command in local mode")
	})

	s.RunForSystemAdminAndLocal("Create outgoing webhook from a specific user", func(c client.Client) {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("team", s.th.BasicTeam.Id, "")
		cmd.Flags().String("user", s.th.BasicUser.Username, "")
		cmd.Flags().String("owner", s.th.BasicUser.Username, "")
		cmd.Flags().String("display-name", model.NewId(), "")
		cmd.Flags().String("channel", s.th.BasicChannel.Id, "")
		cmd.Flags().String("trigger-when", "exact", "")
		cmd.Flags().StringArray("url", []string{"http://example.com"}, "")

		err := createOutgoingWebhookCmdF(s.th.SystemAdminClient, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		webhook := printer.GetLines()[0].(*model.OutgoingWebhook)
		s.Require().Equal(s.th.BasicTeam.Id, webhook.TeamId)
		s.Require().Equal(s.th.BasicUser.Id, webhook.CreatorId)
		s.Require().Equal(s.th.BasicUser.Username, webhook.Username)
	})
}
