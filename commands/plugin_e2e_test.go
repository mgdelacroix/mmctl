// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestPluginListCmdF() {
	s.SetupTestHelper().InitBasic()

	pluginArg := "tmpPlugin"

	s.RunForClient("Error when appropriate permissions are not available", func(c client.Client) {
		printer.Clean()

		s.th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.PluginSettings.Enable = true
		})

		cmd := &cobra.Command{}

		err := pluginListCmdF(c, cmd, []string{pluginArg})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Equal("Unable to list plugins. Error: : You do not have the appropriate permissions., ", err.Error())
	})

	s.RunForSystemAdminAndLocal("Error when plugins are disabled", func(c client.Client) {
		printer.Clean()

		s.th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.PluginSettings.Enable = false
		})

		cmd := &cobra.Command{}

		err := pluginListCmdF(c, cmd, []string{pluginArg})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Equal("Unable to list plugins. Error: : Plugins have been disabled. Please check your logs for details., ", err.Error())
	})

	s.RunForSystemAdminAndLocal("Success when appropriate permissions are available", func(c client.Client) {
		printer.Clean()

		s.th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.PluginSettings.Enable = true
		})

		cmd := &cobra.Command{}

		err := pluginListCmdF(c, cmd, []string{pluginArg})
		s.Require().Nil(err)
	})

	s.RunForSystemAdminAndLocal("Print json plugins", func(c client.Client) {
		printer.Clean()

		s.th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.PluginSettings.Enable = true
		})

		cmd := &cobra.Command{}
		cmd.Flags().String("format", "json", "")

		err := pluginListCmdF(c, cmd, []string{pluginArg})
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Nil(err)
	})

	s.RunForSystemAdminAndLocal("Print the plain enabled and disabled plugins", func(c client.Client) {
		printer.Clean()

		s.th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.PluginSettings.Enable = true
		})

		cmd := &cobra.Command{}

		err := pluginListCmdF(c, cmd, []string{pluginArg})
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Nil(err)
	})
}
