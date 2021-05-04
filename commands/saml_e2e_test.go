// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
)

func (s *MmctlE2ETestSuite) TestSamlAuthDataResetCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()
	defer s.TearDownTest()

	// sanity check
	s.Require().NotNil(s.th.App.Saml())

	user := s.th.BasicUser
	resetAuthDataToID := func() {
		_, err := s.th.App.Srv().Store.User().UpdateAuthData(
			user.Id, model.USER_AUTH_SERVICE_SAML, model.NewString("some-id"), "", false)
		s.Require().Nil(err)
	}
	clearCache := func() {
		err := s.th.App.Srv().InvalidateAllCaches()
		s.Require().Nil(err)
	}
	checkAuthDataWasNotReset := func() {
		clearCache()
		retrievedUser, appErr := s.th.App.GetUser(user.Id)
		s.Require().Nil(appErr)
		s.Require().Equal("some-id", *retrievedUser.AuthData)
	}
	checkAuthDataWasReset := func() {
		clearCache()
		retrievedUser, appErr := s.th.App.GetUser(user.Id)
		s.Require().Nil(appErr)
		s.Require().Equal(user.Email, *retrievedUser.AuthData)
	}

	s.Run("normal client gets permissions error", func() {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("yes", true, "")
		err := samlAuthDataResetCmdF(s.th.Client, cmd, nil)
		s.Require().NotNil(err)
	})

	s.RunForSystemAdminAndLocal("System Admin and Local", func(c client.Client) {
		resetAuthDataToID()
		s.Run("dry run", func() {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("dry-run", true, "")
			err := samlAuthDataResetCmdF(c, cmd, nil)
			s.Require().Nil(err)

			checkAuthDataWasNotReset()
		})

		s.Run("real run", func() {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("yes", true, "")
			err := samlAuthDataResetCmdF(c, cmd, nil)
			s.Require().Nil(err)

			checkAuthDataWasReset()
		})

		resetAuthDataToID()
		s.Run("with specific user IDs", func() {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("yes", true, "")
			cmd.Flags().StringSlice("users", []string{s.th.BasicUser2.Id}, "")
			err := samlAuthDataResetCmdF(c, cmd, nil)
			s.Require().Nil(err)
			checkAuthDataWasNotReset()

			cmd = &cobra.Command{}
			cmd.Flags().Bool("yes", true, "")
			cmd.Flags().StringSlice("users", []string{user.Id}, "")
			err = samlAuthDataResetCmdF(c, cmd, nil)
			s.Require().Nil(err)
			checkAuthDataWasReset()
		})

		resetAuthDataToID()
		// delete user
		deleteUserErr := s.th.App.UpdateUserActive(user.Id, false)
		s.Require().Nil(deleteUserErr)
		s.Run("without deleted users", func() {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("yes", true, "")
			err := samlAuthDataResetCmdF(c, cmd, nil)
			s.Require().Nil(err)

			checkAuthDataWasNotReset()
		})
		s.Run("with deleted users", func() {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("yes", true, "")
			cmd.Flags().Bool("include-deleted", true, "")
			err := samlAuthDataResetCmdF(c, cmd, nil)
			s.Require().Nil(err)

			checkAuthDataWasReset()
		})
		// undelete user
		undeleteUserErr := s.th.App.UpdateUserActive(user.Id, true)
		s.Require().Nil(undeleteUserErr)
	})
}
