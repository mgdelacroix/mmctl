// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestRenameTeamCmdF() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Error renaming team which does not exist", func(c client.Client) {
		printer.Clean()
		nonExistentTeamName := "existingName"
		cmd := &cobra.Command{}
		args := []string{nonExistentTeamName}
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")

		err := renameTeamCmdF(c, cmd, args)
		s.Require().EqualError(err, "Unable to find team 'existingName', to see the all teams try 'team list' command")
	})

	s.RunForSystemAdminAndLocal("Rename an existing team", func(c client.Client) {
		printer.Clean()

		cmd := &cobra.Command{}
		args := []string{s.th.BasicTeam.Name}
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")

		err := renameTeamCmdF(c, cmd, args)
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal("'"+s.th.BasicTeam.Name+"' team renamed", printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Permission error renaming an existing team", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		args := []string{s.th.BasicTeam.Name}
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")

		err := renameTeamCmdF(s.th.Client, cmd, args)
		s.Require().Error(err)
		s.Len(printer.GetLines(), 0)
		s.Equal("Cannot rename team '"+s.th.BasicTeam.Name+"', error : : You do not have the appropriate permissions., ", err.Error())
	})
}

func (s *MmctlE2ETestSuite) TestModifyTeamsCmdF() {
	s.SetupTestHelper().InitBasic()
	s.RunForSystemAdminAndLocal("system & local accounts can set a team to private", func(c client.Client) {
		printer.Clean()
		teamID := s.th.BasicTeam.Id
		cmd := &cobra.Command{}
		cmd.Flags().Bool("private", true, "")
		err := modifyTeamsCmdF(c, cmd, []string{teamID})
		s.Require().NoError(err)

		s.Require().Equal(model.TEAM_INVITE, printer.GetLines()[0].(*model.Team).Type)
		// teardown
		appErr := s.th.App.UpdateTeamPrivacy(teamID, model.TEAM_OPEN, true)
		s.Require().Nil(appErr)
		t, err := s.th.App.GetTeam(teamID)
		s.Require().Nil(err)
		s.th.BasicTeam = t
	})
	s.Run("user that creates the team can't set team's privacy due to permissions", func() {
		printer.Clean()
		teamID := s.th.BasicTeam.Id
		cmd := &cobra.Command{}
		cmd.Flags().Bool("private", true, "")
		err := modifyTeamsCmdF(s.th.Client, cmd, []string{teamID})
		s.Require().NoError(err)
		s.Require().Equal(
			fmt.Sprintf("Unable to modify team '%s' error: : You do not have the appropriate permissions., ", s.th.BasicTeam.Name),
			printer.GetErrorLines()[0],
		)
		t, appErr := s.th.App.GetTeam(teamID)
		s.Require().Nil(appErr)
		s.Require().Equal(model.TEAM_OPEN, t.Type)
	})
	s.Run("basic user with normal permissions that hasn't created the team can't set team's privacy", func() {
		printer.Clean()
		teamID := s.th.BasicTeam.Id
		cmd := &cobra.Command{}
		cmd.Flags().Bool("private", true, "")
		s.th.LoginBasic2()
		err := modifyTeamsCmdF(s.th.Client, cmd, []string{teamID})
		s.Require().NoError(err)
		s.Require().Equal(
			fmt.Sprintf("Unable to modify team '%s' error: : You do not have the appropriate permissions., ", s.th.BasicTeam.Name),
			printer.GetErrorLines()[0],
		)
		t, appErr := s.th.App.GetTeam(teamID)
		s.Require().Nil(appErr)
		s.Require().Equal(model.TEAM_OPEN, t.Type)
	})
}

func (s *MmctlE2ETestSuite) TestArchiveTeamsCmd() {
	s.SetupTestHelper().InitBasic()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("confirm", true, "Confirm you really want to archive the team and a DB backup has been performed.")

	s.RunForAllClients("Archive nonexistent team", func(c client.Client) {
		printer.Clean()

		err := archiveTeamsCmdF(c, cmd, []string{"unknown-team"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to find team 'unknown-team'", printer.GetErrorLines()[0])
	})

	s.RunForSystemAdminAndLocal("Archive basic team", func(c client.Client) {
		printer.Clean()

		err := archiveTeamsCmdF(c, cmd, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		team := printer.GetLines()[0].(*model.Team)
		s.Require().Equal(s.th.BasicTeam.Name, team.Name)
		s.Require().Len(printer.GetErrorLines(), 0)

		basicTeam, err := s.th.App.GetTeam(s.th.BasicTeam.Id)
		s.Require().Nil(err)
		s.Require().NotZero(basicTeam.DeleteAt)

		err = s.th.App.RestoreTeam(s.th.BasicTeam.Id)
		s.Require().Nil(err)
	})

	s.Run("Archive team without permissions", func() {
		printer.Clean()

		err := archiveTeamsCmdF(s.th.Client, cmd, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Contains(printer.GetErrorLines()[0], "You do not have the appropriate permissions.")

		basicTeam, err := s.th.App.GetTeam(s.th.BasicTeam.Id)
		s.Require().Nil(err)
		s.Require().Zero(basicTeam.DeleteAt)
	})
}
