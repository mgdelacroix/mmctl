// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestChannelGroupEnableCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	channelName := api4.GenerateTestChannelName()
	channel, appErr := s.th.App.CreateChannel(&model.Channel{
		TeamId:      s.th.BasicTeam.Id,
		Name:        channelName,
		DisplayName: "dn_" + channelName,
		Type:        model.CHANNEL_OPEN,
	}, false)
	s.Require().Nil(appErr)
	defer func() {
		err := s.th.App.DeleteChannel(channel, "")
		s.Require().Nil(err)
	}()

	id := model.NewId()
	group, appErr := s.th.App.CreateGroup(&model.Group{
		DisplayName: "dn_" + id,
		Name:        model.NewString("name" + id),
		Source:      model.GroupSourceLdap,
		Description: "description_" + id,
		RemoteId:    model.NewId(),
	})
	s.Require().Nil(appErr)
	defer func() {
		_, err := s.th.App.DeleteGroup(group.Id)
		s.Require().Nil(err)
	}()

	_, appErr = s.th.App.UpsertGroupSyncable(&model.GroupSyncable{
		GroupId:    group.Id,
		SyncableId: channel.Id,
		Type:       model.GroupSyncableTypeChannel,
	})
	s.Require().Nil(appErr)
	defer func() {
		_, err := s.th.App.DeleteGroupSyncable(group.Id, channel.Id, model.GroupSyncableTypeChannel)
		s.Require().Nil(err)
	}()

	s.Run("Should not allow regular user to enable group for channel", func() {
		printer.Clean()

		err := channelGroupEnableCmdF(s.th.Client, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Should enable group sync for the channel", func(c client.Client) {
		printer.Clean()

		err := channelGroupEnableCmdF(c, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().NoError(err)

		channel.GroupConstrained = model.NewBool(false)
		defer func() {
			_, err := s.th.App.UpdateChannel(channel)
			s.Require().Nil(err)
		}()

		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		ch, appErr := s.th.App.GetChannel(channel.Id)
		s.Require().Nil(appErr)
		s.Require().True(ch.IsGroupConstrained())
	})
}

func (s *MmctlE2ETestSuite) TestChannelGroupDisableCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	channelName := api4.GenerateTestChannelName()
	channel, appErr := s.th.App.CreateChannel(&model.Channel{
		TeamId:      s.th.BasicTeam.Id,
		Name:        channelName,
		DisplayName: "dn_" + channelName,
		Type:        model.CHANNEL_OPEN,
	}, false)
	s.Require().Nil(appErr)
	defer func() {
		err := s.th.App.DeleteChannel(channel, "")
		s.Require().Nil(err)
	}()

	id := model.NewId()
	group, appErr := s.th.App.CreateGroup(&model.Group{
		DisplayName: "dn_" + id,
		Name:        model.NewString("name" + id),
		Source:      model.GroupSourceLdap,
		Description: "description_" + id,
		RemoteId:    model.NewId(),
	})
	s.Require().Nil(appErr)
	defer func() {
		_, err := s.th.App.DeleteGroup(group.Id)
		s.Require().Nil(err)
	}()

	_, appErr = s.th.App.UpsertGroupSyncable(&model.GroupSyncable{
		GroupId:    group.Id,
		SyncableId: channel.Id,
		Type:       model.GroupSyncableTypeChannel,
	})
	s.Require().Nil(appErr)
	defer func() {
		_, err := s.th.App.DeleteGroupSyncable(group.Id, channel.Id, model.GroupSyncableTypeChannel)
		s.Require().Nil(err)
	}()

	channel.GroupConstrained = model.NewBool(true)
	defer func() {
		_, err := s.th.App.UpdateChannel(channel)
		s.Require().Nil(err)
	}()

	s.Run("Should not allow regular user to disable group for channel", func() {
		printer.Clean()

		err := channelGroupDisableCmdF(s.th.Client, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Should disable group sync for the channel", func(c client.Client) {
		printer.Clean()

		err := channelGroupDisableCmdF(c, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().NoError(err)

		channel.GroupConstrained = model.NewBool(true)
		defer func() {
			_, err := s.th.App.UpdateChannel(channel)
			s.Require().Nil(err)
		}()

		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		ch, appErr := s.th.App.GetChannel(channel.Id)
		s.Require().Nil(appErr)
		s.Require().False(ch.IsGroupConstrained())
	})
}

func (s *MmctlE2ETestSuite) TestListLdapGroupsCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()
	configForLdap(s.th)

	s.Run("MM-T3977 Should not allow regular user to list LDAP groups", func() {
		printer.Clean()

		err := listLdapGroupsCmdF(s.th.Client, &cobra.Command{}, nil)
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("MM-T3976 Should list LDAP groups", func(c client.Client) {
		printer.Clean()

		// we rely on the test data generated for the openldap server
		// i.e. "test-data.ldif" script
		err := listLdapGroupsCmdF(c, &cobra.Command{}, nil)
		s.Require().NoError(err)
		s.Require().NotEmpty(printer.GetLines())
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlE2ETestSuite) TestChannelGroupStatusCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	channelName := api4.GenerateTestChannelName()
	channel, appErr := s.th.App.CreateChannel(&model.Channel{
		TeamId:           s.th.BasicTeam.Id,
		Name:             channelName,
		DisplayName:      "dn_" + channelName,
		Type:             model.CHANNEL_OPEN,
		GroupConstrained: model.NewBool(true),
	}, false)
	s.Require().Nil(appErr)
	defer func() {
		err := s.th.App.DeleteChannel(channel, "")
		s.Require().Nil(err)
	}()

	channelName2 := api4.GenerateTestChannelName()
	channel2, appErr := s.th.App.CreateChannel(&model.Channel{
		TeamId:      s.th.BasicTeam.Id,
		Name:        channelName2,
		DisplayName: "dn_" + channelName2,
		Type:        model.CHANNEL_OPEN,
	}, false)
	s.Require().Nil(appErr)
	defer func() {
		err := s.th.App.DeleteChannel(channel2, "")
		s.Require().Nil(err)
	}()

	s.RunForAllClients("MM-T3974 Should allow to get status of a group constrained channel", func(c client.Client) {
		printer.Clean()

		err := channelGroupStatusCmdF(c, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().NoError(err)

		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Enabled")
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForAllClients("MM-T3975 Should allow to get status of a regular channel", func(c client.Client) {
		printer.Clean()

		err := channelGroupStatusCmdF(c, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName2})
		s.Require().NoError(err)

		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Disabled")
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlE2ETestSuite) TestChannelGroupListCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	channelName := api4.GenerateTestChannelName()
	channel, appErr := s.th.App.CreateChannel(&model.Channel{
		TeamId:      s.th.BasicTeam.Id,
		Name:        channelName,
		DisplayName: "dn_" + channelName,
		Type:        model.CHANNEL_OPEN,
	}, false)
	s.Require().Nil(appErr)
	defer func() {
		err := s.th.App.DeleteChannel(channel, "")
		s.Require().Nil(err)
	}()

	id := model.NewId()
	group, appErr := s.th.App.CreateGroup(&model.Group{
		DisplayName: "dn_" + id,
		Name:        model.NewString("name" + id),
		Source:      model.GroupSourceLdap,
		Description: "description_" + id,
		RemoteId:    model.NewId(),
	})
	s.Require().Nil(appErr)
	defer func() {
		_, err := s.th.App.DeleteGroup(group.Id)
		s.Require().Nil(err)
	}()

	_, appErr = s.th.App.UpsertGroupSyncable(&model.GroupSyncable{
		GroupId:    group.Id,
		SyncableId: channel.Id,
		Type:       model.GroupSyncableTypeChannel,
	})
	s.Require().Nil(appErr)
	defer func() {
		_, err := s.th.App.DeleteGroupSyncable(group.Id, channel.Id, model.GroupSyncableTypeChannel)
		s.Require().Nil(err)
	}()

	s.Run("MM-T3970 Should not allow regular user to get list of LDAP groups in a channel", func() {
		printer.Clean()

		err := channelGroupListCmdF(s.th.Client, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("MM-T3969 Should allow to get list of LDAP groups in a channel", func(c client.Client) {
		printer.Clean()

		err := channelGroupListCmdF(c, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().NoError(err)

		s.Require().Len(printer.GetLines(), 1)
		gs, ok := printer.GetLines()[0].(*model.GroupWithSchemeAdmin)
		s.Require().True(ok)
		s.Require().Equal(gs.Group, *group)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlE2ETestSuite) TestTeamGroupDisableCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	team, _, cleanUpFn := createTestGroupTeam(s)
	defer cleanUpFn()

	team.GroupConstrained = model.NewBool(true)
	_, err := s.th.App.UpdateTeam(team)
	s.Require().Nil(err)

	s.Run("MM-T3919 Should not allow regular user to disable group for team", func() {
		printer.Clean()

		err := teamGroupDisableCmdF(s.th.Client, &cobra.Command{}, []string{team.Name})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("MM-T3920 Should disable group sync for the team", func(c client.Client) {
		printer.Clean()

		err := teamGroupDisableCmdF(c, &cobra.Command{}, []string{team.Name})
		s.Require().NoError(err)

		team.GroupConstrained = model.NewBool(true)
		defer func() {
			_, err := s.th.App.UpdateTeam(team)
			s.Require().Nil(err)
		}()

		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		tm, appErr := s.th.App.GetTeam(team.Id)
		s.Require().Nil(appErr)
		s.Require().False(tm.IsGroupConstrained())
	})
}

func (s *MmctlE2ETestSuite) TestTeamGroupEnableCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	team, _, cleanUpFn := createTestGroupTeam(s)
	defer cleanUpFn()

	s.Run("MM-T3917 Should not allow regular user to enable group for team", func() {
		printer.Clean()

		err := teamGroupEnableCmdF(s.th.Client, &cobra.Command{}, []string{team.Name})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("MM-T3918 Should enable group sync for the team", func(c client.Client) {
		printer.Clean()

		err := teamGroupEnableCmdF(c, &cobra.Command{}, []string{team.Name})
		s.Require().NoError(err)

		team.GroupConstrained = model.NewBool(false)
		defer func() {
			_, err := s.th.App.UpdateTeam(team)
			s.Require().Nil(err)
		}()

		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		tm, appErr := s.th.App.GetTeam(team.Id)
		s.Require().Nil(appErr)
		s.Require().True(tm.IsGroupConstrained())
	})
}

func (s *MmctlE2ETestSuite) TestTeamGroupStatusCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	team, _, cleanUpFn := createTestGroupTeam(s)
	defer func() {
		cleanUpFn()
	}()

	teamName2 := api4.GenerateTestTeamName()
	team2, appErr := s.th.App.CreateTeam(&model.Team{
		Name:        teamName2,
		DisplayName: "dn_" + teamName2,
		Type:        model.TEAM_INVITE,
	})
	s.Require().Nil(appErr)
	defer func() {
		err := s.th.App.PermanentDeleteTeam(team2)
		s.Require().Nil(err)
	}()

	s.Run("MM-T3921 Should not allow regular user to get status of LDAP groups in a team where they are not a member of", func() {
		printer.Clean()

		err := teamGroupStatusCmdF(s.th.Client, &cobra.Command{}, []string{team.Name})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	_, _, appErr = s.th.App.AddUserToTeam(team.Id, s.th.BasicUser.Id, s.th.SystemAdminUser.Id)
	s.Require().Nil(appErr)

	_, _, appErr = s.th.App.AddUserToTeam(team2.Id, s.th.BasicUser.Id, s.th.SystemAdminUser.Id)
	s.Require().Nil(appErr)

	s.RunForAllClients("MM-T3922 Should allow to get status of a group constrained team", func(c client.Client) {
		printer.Clean()

		err := teamGroupStatusCmdF(c, &cobra.Command{}, []string{team.Name})
		s.Require().NoError(err)

		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Enabled")
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForAllClients("MM-T3923 Should allow to get status of a regular team", func(c client.Client) {
		printer.Clean()

		err := teamGroupStatusCmdF(c, &cobra.Command{}, []string{teamName2})
		s.Require().NoError(err)

		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Disabled")
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlE2ETestSuite) TestTeamGroupListCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	team, group, cleanUpFn := createTestGroupTeam(s)
	defer func() {
		cleanUpFn()
	}()

	s.Run("MM-T3924 Should not allow regular user to get list of LDAP groups in a team", func() {
		printer.Clean()

		err := teamGroupListCmdF(s.th.Client, &cobra.Command{}, []string{team.Name})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("MM-T3925 Should allow to get list of LDAP groups in a team", func(c client.Client) {
		printer.Clean()

		err := teamGroupListCmdF(c, &cobra.Command{}, []string{team.Name})
		s.Require().NoError(err)

		s.Require().Len(printer.GetLines(), 1)
		gs, ok := printer.GetLines()[0].(*model.GroupWithSchemeAdmin)
		s.Require().True(ok)
		s.Require().Equal(gs.Group, *group)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func createTestGroupTeam(s *MmctlE2ETestSuite) (*model.Team, *model.Group, func()) {
	teamName := api4.GenerateTestTeamName()
	team, appErr := s.th.App.CreateTeam(&model.Team{
		Name:             teamName,
		DisplayName:      "dn_" + teamName,
		Type:             model.TEAM_OPEN,
		GroupConstrained: model.NewBool(true),
	})
	s.Require().Nil(appErr)

	id := model.NewId()
	group, appErr := s.th.App.CreateGroup(&model.Group{
		DisplayName: "dn_" + id,
		Name:        model.NewString("name" + id),
		Source:      model.GroupSourceLdap,
		Description: "description_" + id,
		RemoteId:    model.NewId(),
	})
	s.Require().Nil(appErr)

	_, appErr = s.th.App.UpsertGroupSyncable(&model.GroupSyncable{
		GroupId:    group.Id,
		SyncableId: team.Id,
		Type:       model.GroupSyncableTypeTeam,
	})
	s.Require().Nil(appErr)

	cleanUpFn := func() {
		_, err := s.th.App.DeleteGroupSyncable(group.Id, team.Id, model.GroupSyncableTypeTeam)
		s.Require().Nil(err)

		_, err = s.th.App.DeleteGroup(group.Id)
		s.Require().Nil(err)

		err = s.th.App.PermanentDeleteTeamId(team.Id)
		s.Require().Nil(err)
	}

	return team, group, cleanUpFn
}
