// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestUserActivateCmd() {
	s.Run("Activate user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, true).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userActivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Try to activate unexistent user", func() {
		printer.Clean()
		emailArg := "example@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		err := userActivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* user %s not found\n\n", emailArg), printer.GetErrorLines()[0])
	})

	s.Run("Fail to activate user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, true).
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := userActivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Errorf("unable to change activation status of user: %v", mockUser.Id).Error(), printer.GetErrorLines()[0])
	})

	s.Run("Activate several users with unexistent ones and failed ones", func() {
		printer.Clean()
		emailArgs := []string{"example0@example0.com", "null", "example2@example2.com", "failure@failure.com", "example4@example4.com"}
		mockUser0 := model.User{Id: "example0", Username: "ExampleUser0", Email: emailArgs[0]}
		mockUser2 := model.User{Id: "example2", AuthService: "other", Username: "ExampleUser2", Email: emailArgs[2]}
		mockUser3 := model.User{Id: "failure", Username: "FailureUser", Email: emailArgs[3]}
		mockUser4 := model.User{Id: "example4", Username: "ExampleUser4", Email: emailArgs[4]}

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[0], "").
			Return(&mockUser0, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[2], "").
			Return(&mockUser2, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[3], "").
			Return(&mockUser3, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[4], "").
			Return(&mockUser4, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser0.Id, true).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser2.Id, true).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser3.Id, true).
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser4.Id, true).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userActivateCmdF(s.client, &cobra.Command{}, emailArgs)
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* user %s not found\n\n", emailArgs[1]), printer.GetErrorLines()[0])
		s.Require().Equal(fmt.Sprintf("unable to change activation status of user: %v", mockUser3.Id), printer.GetErrorLines()[1])
	})
}

func (s *MmctlUnitTestSuite) TestDeactivateUserCmd() {
	s.Run("Deactivate user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Try to deactivate unexistent user", func() {
		printer.Clean()
		emailArg := "example@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusBadRequest}}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* user %v not found\n\n", emailArg), printer.GetErrorLines()[0])
	})

	s.Run("Fail to deactivate user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Errorf("unable to change activation status of user: %v", mockUser.Id).Error(), printer.GetErrorLines()[0])
	})

	s.Run("Deactivate SSO user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", AuthService: "other", Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("You must also deactivate user "+mockUser.Id+" in the SSO provider or they will be reactivated on next login or sync.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Deactivate several users with unexistent ones, SSO ones and failed ones", func() {
		printer.Clean()
		emailArgs := []string{"example0@example0.com", "null", "example2@example2.com", "failure@failure.com", "example4@example4.com"}
		mockUser0 := model.User{Id: "example0", Username: "ExampleUser0", Email: emailArgs[0]}
		mockUser2 := model.User{Id: "example2", AuthService: "other", Username: "ExampleUser2", Email: emailArgs[2]}
		mockUser3 := model.User{Id: "failure", Username: "FailureUser", Email: emailArgs[3]}
		mockUser4 := model.User{Id: "example4", Username: "ExampleUser4", Email: emailArgs[4]}

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[0], "").
			Return(&mockUser0, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[2], "").
			Return(&mockUser2, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[3], "").
			Return(&mockUser3, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[4], "").
			Return(&mockUser4, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser0.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser2.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser3.Id, false).
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser4.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, emailArgs)
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("You must also deactivate user "+mockUser2.Id+" in the SSO provider or they will be reactivated on next login or sync.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* user %v not found\n\n", emailArgs[1]), printer.GetErrorLines()[0])
		s.Require().Equal(fmt.Errorf("unable to change activation status of user: %v", mockUser3.Id).Error(), printer.GetErrorLines()[1])
	})
}

func (s *MmctlUnitTestSuite) TestDeleteUsersCmd() {
	email1 := "user1@example.com"
	email2 := "user2@example.com"
	userID1 := model.NewId()
	userID2 := model.NewId()
	mockUser1 := model.User{Username: "User1", Email: email1, Id: userID1}
	mockUser2 := model.User{Username: "User2", Email: email2, Id: userID2}

	s.Run("Delete users with confirm false returns an error", func() {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", false, "")
		err := deleteUsersCmdF(s.client, cmd, []string{"some"})
		s.Require().NotNil(err)
		s.Require().Equal(err.Error(), "aborted: You did not answer YES exactly, in all capitals")
	})

	s.Run("Delete user that does not exist in db returns an error", func() {
		printer.Clean()
		arg := "userdoesnotexist@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")
		err := deleteUsersCmdF(s.client, cmd, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* user %s not found\n\n", arg), printer.GetErrorLines()[0])
	})

	s.Run("Delete users should delete users", func() {
		printer.Clean()

		s.client.
			EXPECT().
			GetUserByEmail(email1, "").
			Return(&mockUser1, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PermanentDeleteUser(userID1).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(email2, "").
			Return(&mockUser2, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PermanentDeleteUser(userID2).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		err := deleteUsersCmdF(s.client, cmd, []string{email1, email2})
		s.Require().Nil(err)
		s.Require().Equal(&mockUser1, printer.GetLines()[0])
		s.Require().Equal(&mockUser2, printer.GetLines()[1])
	})

	s.Run("Delete users with error on PermanentDeleteUser returns an error", func() {
		printer.Clean()

		mockError := &model.AppError{
			Message:       "An error occurred on deleting a user",
			DetailedError: "User cannot be deleted",
			Where:         "User.deleteUser",
		}

		s.client.
			EXPECT().
			GetUserByEmail(email1, "").
			Return(&mockUser1, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PermanentDeleteUser(userID1).
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		err := deleteUsersCmdF(s.client, cmd, []string{email1})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to delete user 'User1' error: User.deleteUser: An error occurred on deleting a user, User cannot be deleted",
			printer.GetErrorLines()[0])
	})

	s.Run("Delete two users, first fails with error other passes", func() {
		printer.Clean()

		mockError := &model.AppError{
			Message:       "An error occurred on deleting a user",
			DetailedError: "User cannot be deleted",
			Where:         "User.deleteUser",
		}

		s.client.
			EXPECT().
			GetUserByEmail(email1, "").
			Return(&mockUser1, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(email2, "").
			Return(&mockUser2, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PermanentDeleteUser(userID1).
			Return(false, &model.Response{Error: mockError}).
			Times(1)
		s.client.
			EXPECT().
			PermanentDeleteUser(userID2).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		err := deleteUsersCmdF(s.client, cmd, []string{email1, email2})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(&mockUser2, printer.GetLines()[0])
		s.Require().Equal("Unable to delete user 'User1' error: User.deleteUser: An error occurred on deleting a user, User cannot be deleted",
			printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestDeleteAllUsersCmd() {
	s.Run("Delete all users", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		s.client.
			EXPECT().
			PermanentDeleteAllUsers().
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := deleteAllUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(printer.GetLines()[0], "All users successfully deleted")
	})

	s.Run("Delete all users call fails", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		s.client.
			EXPECT().
			PermanentDeleteAllUsers().
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := deleteAllUsersCmdF(s.client, cmd, []string{})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestSearchUserCmd() {
	s.Run("Search for an existing user", func() {
		emailArg := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		err := searchUserCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().Nil(err)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Search for a nonexistent user", func() {
		printer.Clean()
		arg := "example@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := searchUserCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Equal("1 error occurred:\n\t* user example@example.com not found\n\n", printer.GetErrorLines()[0])
	})

	s.Run("Avoid path traversal", func() {
		printer.Clean()
		arg := "test/../hello?@mattermost.com"

		err := searchUserCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Equal("1 error occurred:\n\t* user test/../hello?@mattermost.com not found\n\n", printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestChangePasswordUserCmdF() {
	s.Run("Change password for oneself", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "userId", Username: "ExampleUser", Email: emailArg}
		currentPassword := "current-password"
		password := "password"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserPassword(mockUser.Id, currentPassword, password).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().String("password", password, "")
		cmd.Flags().String("current", currentPassword, "")

		err := changePasswordUserCmdF(s.client, cmd, []string{emailArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Change password for another user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "userId", Username: "ExampleUser", Email: emailArg}
		password := "password"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserPassword(mockUser.Id, "", password).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().String("password", password, "")

		err := changePasswordUserCmdF(s.client, cmd, []string{emailArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Error when changing password for oneself", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "userId", Username: "ExampleUser", Email: emailArg}
		mockError := model.AppError{Message: "Mock error"}
		currentPassword := "current-password"
		password := "password"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserPassword(mockUser.Id, currentPassword, password).
			Return(true, &model.Response{Error: &mockError}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().String("password", password, "")
		cmd.Flags().String("current", currentPassword, "")

		err := changePasswordUserCmdF(s.client, cmd, []string{emailArg})
		s.Require().Error(err)
		s.Require().EqualError(err, "changing user password failed: : Mock error, ")
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Error when changing password for another user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "userId", Username: "ExampleUser", Email: emailArg}
		mockError := model.AppError{Message: "Mock error"}
		password := "password"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserPassword(mockUser.Id, "", password).
			Return(true, &model.Response{Error: &mockError}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().String("password", password, "")

		err := changePasswordUserCmdF(s.client, cmd, []string{emailArg})
		s.Require().Error(err)
		s.Require().EqualError(err, "changing user password failed: : Mock error, ")
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Error changing password for a nonexisting user", func() {
		printer.Clean()
		arg := "example@example.com"
		password := "password"

		s.client.
			EXPECT().
			GetUserByEmail(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().String("password", password, "")

		err := changePasswordUserCmdF(s.client, cmd, []string{arg})
		s.Require().Error(err)
		s.Require().EqualError(err, "user example@example.com not found")
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Change password by a hashed one", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "userId", Username: "ExampleUser", Email: emailArg}
		hashedPassword := "hashed-password"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserHashedPassword(mockUser.Id, hashedPassword).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().String("password", hashedPassword, "")
		cmd.Flags().Bool("hashed", true, "")

		err := changePasswordUserCmdF(s.client, cmd, []string{emailArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Error when changing password by a hashed one", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "userId", Username: "ExampleUser", Email: emailArg}
		mockError := model.AppError{Message: "Mock error"}
		hashedPassword := "hashed-password"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserHashedPassword(mockUser.Id, hashedPassword).
			Return(true, &model.Response{Error: &mockError}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().String("password", hashedPassword, "")
		cmd.Flags().Bool("hashed", true, "")

		err := changePasswordUserCmdF(s.client, cmd, []string{emailArg})
		s.Require().EqualError(err, "changing user hashed password failed: : Mock error, ")
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestSendPasswordResetEmailCmd() {
	s.Run("Send one reset email", func() {
		printer.Clean()
		emailArg := "example@example.com"

		s.client.
			EXPECT().
			SendPasswordResetEmail(emailArg).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := sendPasswordResetEmailCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Send one reset email and receive error", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			SendPasswordResetEmail(emailArg).
			Return(false, &model.Response{Error: &mockError}).
			Times(1)

		err := sendPasswordResetEmailCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable send reset password email to email "+emailArg+". Error: "+mockError.Error(), printer.GetErrorLines()[0])
	})

	s.Run("Send several reset emails and receive some errors", func() {
		printer.Clean()
		emailArg := []string{
			"example1@example.com",
			"error1@example.com",
			"error2@example.com",
			"example2@example.com",
			"example3@example.com"}
		mockError := model.AppError{Id: "Mock Error"}

		for _, email := range emailArg {
			if strings.HasPrefix(email, "error") {
				s.client.
					EXPECT().
					SendPasswordResetEmail(email).
					Return(false, &model.Response{Error: &mockError}).
					Times(1)
			} else {
				s.client.
					EXPECT().
					SendPasswordResetEmail(email).
					Return(false, &model.Response{Error: nil}).
					Times(1)
			}
		}

		err := sendPasswordResetEmailCmdF(s.client, &cobra.Command{}, emailArg)
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal("Unable send reset password email to email "+emailArg[1]+". Error: "+mockError.Error(), printer.GetErrorLines()[0])
		s.Require().Equal("Unable send reset password email to email "+emailArg[2]+". Error: "+mockError.Error(), printer.GetErrorLines()[1])
	})
}

func (s *MmctlUnitTestSuite) TestUserInviteCmd() {
	s.Run("Invite user to an existing team by Id", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := "teamId"

		s.client.
			EXPECT().
			GetTeam(argTeam, "").
			Return(&model.Team{Id: argTeam}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(argTeam, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := userInviteCmdF(s.client, &cobra.Command{}, []string{argUser, argTeam})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("Invites may or may not have been sent.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Invite user to an existing team by name", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := "teamName"
		resultID := "teamId"

		s.client.
			EXPECT().
			GetTeam(argTeam, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(argTeam, "").
			Return(&model.Team{Id: resultID}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(resultID, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := userInviteCmdF(s.client, &cobra.Command{}, []string{argUser, argTeam})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("Invites may or may not have been sent.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Invite user to several existing teams by name and id", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := []string{"teamName1", "teamId2", "teamId3", "teamName4"}
		resultTeamModels := [4]*model.Team{
			{Id: "teamId1"},
			{Id: "teamId2"},
			{Id: "teamId3"},
			{Id: "teamId4"},
		}

		// Setup GetTeam
		s.client.
			EXPECT().
			GetTeam(argTeam[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[1], "").
			Return(resultTeamModels[1], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[2], "").
			Return(resultTeamModels[2], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[3], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		// Setup GetTeamByName
		s.client.
			EXPECT().
			GetTeamByName(argTeam[0], "").
			Return(resultTeamModels[0], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(argTeam[3], "").
			Return(resultTeamModels[3], &model.Response{Error: nil}).
			Times(1)

		// Setup InviteUsersToTeam
		for _, resultTeamModel := range resultTeamModels {
			s.client.
				EXPECT().
				InviteUsersToTeam(resultTeamModel.Id, []string{argUser}).
				Return(false, &model.Response{Error: nil}).
				Times(1)
		}

		err := userInviteCmdF(s.client, &cobra.Command{}, append([]string{argUser}, argTeam...))
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), len(argTeam))
		for i := 0; i < len(argTeam); i++ {
			s.Require().Equal("Invites may or may not have been sent.", printer.GetLines()[i])
		}
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Invite user to an un-existing team", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := "unexistent"

		s.client.
			EXPECT().
			GetTeam(argTeam, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(argTeam, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := userInviteCmdF(s.client, &cobra.Command{}, []string{argUser, argTeam})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("can't find team '"+argTeam+"'", printer.GetErrorLines()[0])
	})

	s.Run("Invite user to an existing team and fail invite", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := "teamId"
		resultName := "teamName"
		mockError := model.NewAppError("", "Mock Error", nil, "", 0)

		s.client.
			EXPECT().
			GetTeam(argTeam, "").
			Return(&model.Team{Id: argTeam, Name: resultName}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(argTeam, []string{argUser}).
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		err := userInviteCmdF(s.client, &cobra.Command{}, []string{argUser, argTeam})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to invite user with email "+argUser+" to team "+resultName+". Error: "+mockError.Error(), printer.GetErrorLines()[0])
	})

	s.Run("Invite user to several existing and non-existing teams by name and id and reject one invite", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := []string{"teamName1", "unexistent", "teamId3", "teamName4", "reject", "teamId6"}
		resultTeamModels := [6]*model.Team{
			{Id: "teamId1", Name: "teamName1"},
			nil,
			{Id: "teamId3", Name: "teamName3"},
			{Id: "teamId4", Name: "teamName4"},
			{Id: "reject", Name: "rejectName"},
			{Id: "teamId6", Name: "teamName6"},
		}
		mockError := model.NewAppError("", "Mock Error", nil, "", 0)

		// Setup GetTeam
		s.client.
			EXPECT().
			GetTeam(argTeam[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[1], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[2], "").
			Return(resultTeamModels[2], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[3], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[4], "").
			Return(resultTeamModels[4], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[5], "").
			Return(resultTeamModels[5], &model.Response{Error: nil}).
			Times(1)

		// Setup GetTeamByName
		s.client.
			EXPECT().
			GetTeamByName(argTeam[0], "").
			Return(resultTeamModels[0], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(argTeam[1], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(argTeam[3], "").
			Return(resultTeamModels[3], &model.Response{Error: nil}).
			Times(1)

		// Setup InviteUsersToTeam
		s.client.
			EXPECT().
			InviteUsersToTeam(resultTeamModels[0].Id, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(resultTeamModels[2].Id, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(resultTeamModels[3].Id, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(resultTeamModels[4].Id, []string{argUser}).
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(resultTeamModels[5].Id, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := userInviteCmdF(s.client, &cobra.Command{}, append([]string{argUser}, argTeam...))
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 4)
		for i := 0; i < 4; i++ {
			s.Require().Equal("Invites may or may not have been sent.", printer.GetLines()[i])
		}
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal("can't find team '"+argTeam[1]+"'", printer.GetErrorLines()[0])
		s.Require().Equal("Unable to invite user with email "+argUser+" to team "+resultTeamModels[4].Name+". Error: "+mockError.Error(), printer.GetErrorLines()[1])
	})
}

func (s *MmctlUnitTestSuite) TestUserCreateCmd() {
	mockUser := model.User{
		Username: "username",
		Password: "password",
		Email:    "email",
	}

	s.Run("Create user with email missing", func() {
		printer.Clean()

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("password", mockUser.Password, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Equal("Email is required: flag accessed but not defined: email", error.Error())
	})

	s.Run("Create user with username missing", func() {
		printer.Clean()

		command := cobra.Command{}
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Equal("Username is required: flag accessed but not defined: username", error.Error())
	})

	s.Run("Create user with password missing", func() {
		printer.Clean()

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Equal("Password is required: flag accessed but not defined: password", error.Error())
	})

	s.Run("Create a regular user", func() {
		printer.Clean()

		s.client.
			EXPECT().
			CreateUser(&mockUser).
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Nil(error)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Create a regular user with welcome email disabled", func() {
		printer.Clean()

		oldDisableWelcomeEmail := mockUser.DisableWelcomeEmail
		mockUser.DisableWelcomeEmail = true
		defer func() { mockUser.DisableWelcomeEmail = oldDisableWelcomeEmail }()

		s.client.
			EXPECT().
			CreateUser(&mockUser).
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")
		command.Flags().Bool("disable-welcome-email", mockUser.DisableWelcomeEmail, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Nil(error)
		printerLines := printer.GetLines()[0]
		printedUser := printerLines.(*model.User)

		s.Require().Equal(&mockUser, printerLines)
		s.Require().Equal(true, printedUser.DisableWelcomeEmail)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Create a regular user with client returning error", func() {
		printer.Clean()

		s.client.
			EXPECT().
			CreateUser(&mockUser).
			Return(&mockUser, &model.Response{Error: &model.AppError{Message: "Remote error"}}).
			Times(1)

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Equal("Unable to create user. Error: : Remote error, ", error.Error())
	})

	s.Run("Create a sysAdmin user", func() {
		printer.Clean()

		s.client.
			EXPECT().
			CreateUser(&mockUser).
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(mockUser.Id, "system_user system_admin").
			Return(true, &model.Response{Error: nil}).
			Times(1)

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")
		command.Flags().Bool("system_admin", true, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Nil(error)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Create a guest user", func() {
		printer.Clean()

		s.client.
			EXPECT().
			CreateUser(&mockUser).
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DemoteUserToGuest(mockUser.Id).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")
		command.Flags().Bool("guest", true, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Nil(error)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Create a sysAdmin user with client returning error", func() {
		printer.Clean()

		s.client.
			EXPECT().
			CreateUser(&mockUser).
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(mockUser.Id, "system_user system_admin").
			Return(false, &model.Response{Error: &model.AppError{Message: "Remote error"}}).
			Times(1)

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")
		command.Flags().Bool("system_admin", true, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Equal("Unable to update user roles. Error: : Remote error, ", error.Error())
	})
}

func (s *MmctlUnitTestSuite) TestUpdateUserEmailCmd() {
	s.Run("Two arguments are not provided", func() {
		printer.Clean()

		command := cobra.Command{}

		error := updateUserEmailCmdF(s.client, &command, []string{})

		s.Require().EqualError(error, "expected two arguments. See help text for details")
	})

	s.Run("Invalid email provided", func() {
		printer.Clean()

		userArg := "testUser"
		emailArg := "invalidEmail"
		command := cobra.Command{}

		error := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().EqualError(error, "invalid email: 'invalidEmail'")
	})

	s.Run("User not found using email, username or id as identifier", func() {
		printer.Clean()

		command := cobra.Command{}
		userArg := "testUser"
		emailArg := "example@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given email", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given username", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given id", StatusCode: http.StatusNotFound}}).
			Times(1)

		err := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().EqualError(err, "user testUser not found")
	})

	s.Run("Client returning error while updating user", func() {
		printer.Clean()

		command := cobra.Command{}
		userArg := "testUser"
		emailArg := "example@example.com"

		currentUser := model.User{Username: "testUser", Password: "password", Email: "email"}

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given email", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userArg, "").
			Return(&currentUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUser(&currentUser).
			Return(nil, &model.Response{Error: &model.AppError{Message: "Remote error"}}).
			Times(1)

		err := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().EqualError(err, ": Remote error, ")
	})

	s.Run("User email is updated successfully using username as identifier", func() {
		printer.Clean()

		command := cobra.Command{}
		userArg := "testUser"
		emailArg := "example@example.com"

		currentUser := model.User{Username: "testUser", Password: "password", Email: "email"}
		updatedUser := model.User{Username: "testUser", Password: "password", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given email", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userArg, "").
			Return(&currentUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUser(&currentUser).
			Return(&updatedUser, &model.Response{Error: nil}).
			Times(1)

		err := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().Nil(err)
		s.Require().Equal(&updatedUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("User email is updated successfully using email as identifier", func() {
		printer.Clean()

		command := cobra.Command{}
		userArg := "user@email.com"
		emailArg := "example@example.com"

		currentUser := model.User{Username: "testUser", Password: "password", Email: "email"}
		updatedUser := model.User{Username: "testUser", Password: "password", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(&currentUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUser(&currentUser).
			Return(&updatedUser, &model.Response{Error: nil}).
			Times(1)

		error := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().Nil(error)
		s.Require().Equal(&updatedUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("User email is updated successfully using id as identifier", func() {
		printer.Clean()

		command := cobra.Command{}
		userArg := "userId"
		emailArg := "example@example.com"

		currentUser := model.User{Username: "testUser", Password: "password", Email: "email"}
		updatedUser := model.User{Username: "testUser", Password: "password", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given email", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given username", StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(userArg, "").
			Return(&currentUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUser(&currentUser).
			Return(&updatedUser, &model.Response{Error: nil}).
			Times(1)

		err := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().Nil(err)
		s.Require().Equal(&updatedUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestResetUserMfaCmd() {
	s.Run("One user without problems", func() {
		printer.Clean()

		s.client.
			EXPECT().
			GetUserByEmail("userId", "").
			Return(&model.User{Id: "userId"}, nil).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserMfa("userId", "", false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := resetUserMfaCmdF(s.client, &cobra.Command{}, []string{"userId"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Cannot find one user", func() {
		printer.Clean()

		s.client.
			EXPECT().
			GetUserByEmail("userId", "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername("userId", "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUser("userId", "").
			Return(nil, nil).
			Times(1)

		err := resetUserMfaCmdF(s.client, &cobra.Command{}, []string{"userId"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "1 error occurred:\n\t* user userId not found\n\n")
	})

	s.Run("One user, unable to reset", func() {
		printer.Clean()
		mockError := model.AppError{Message: "Mock error"}

		s.client.
			EXPECT().
			GetUserByEmail("userId", "").
			Return(&model.User{Id: "userId"}, nil).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserMfa("userId", "", false).
			Return(false, &model.Response{Error: &mockError}).
			Times(1)

		err := resetUserMfaCmdF(s.client, &cobra.Command{}, []string{"userId"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to reset user 'userId' MFA. Error: "+mockError.Error())
	})

	s.Run("Several users, with unknown users and users unable to be reset", func() {
		printer.Clean()
		users := []string{"user0", "error1", "user2", "notfounduser", "user4"}
		mockError := model.AppError{Message: "Mock error"}

		for _, user := range users {
			if user == "notfounduser" {
				s.client.
					EXPECT().
					GetUserByEmail(user, "").
					Return(nil, nil).
					Times(1)

				s.client.
					EXPECT().
					GetUserByUsername(user, "").
					Return(nil, nil).
					Times(1)

				s.client.
					EXPECT().
					GetUser(user, "").
					Return(nil, nil).
					Times(1)
			} else {
				s.client.
					EXPECT().
					GetUserByEmail(user, "").
					Return(&model.User{Id: user}, nil).
					Times(1)
			}
		}

		for _, user := range users {
			if user == "error1" {
				s.client.
					EXPECT().
					UpdateUserMfa(user, "", false).
					Return(false, &model.Response{Error: &mockError}).
					Times(1)
			} else if user != "notfounduser" {
				s.client.
					EXPECT().
					UpdateUserMfa(user, "", false).
					Return(true, &model.Response{Error: nil}).
					Times(1)
			}
		}

		err := resetUserMfaCmdF(s.client, &cobra.Command{}, users)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* user %s not found\n\n", users[3]), printer.GetErrorLines()[0])
		s.Require().Equal("Unable to reset user '"+users[1]+"' MFA. Error: "+mockError.Error(), printer.GetErrorLines()[1])
	})
}

func (s *MmctlUnitTestSuite) TestListUserCmdF() {
	cmd := &cobra.Command{}
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().Int("per-page", 200, "")
	cmd.Flags().Bool("all", false, "")
	cmd.Flags().String("team", "", "")

	s.Run("Listing users with paging", func() {
		printer.Clean()

		email := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: email}

		page := 0
		perPage := 1
		showAll := false
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))

		s.client.
			EXPECT().
			GetUsers(page, perPage, "").
			Return([]*model.User{&mockUser}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
	})

	s.Run("Listing all the users", func() {
		printer.Clean()

		email1 := "example1@example.com"
		mockUser1 := model.User{Username: "ExampleUser1", Email: email1}
		email2 := "example2@example.com"
		mockUser2 := model.User{Username: "ExampleUser2", Email: email2}

		page := 0
		perPage := 1
		showAll := true
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))

		s.client.
			EXPECT().
			GetUsers(0, perPage, "").
			Return([]*model.User{&mockUser1}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUsers(1, perPage, "").
			Return([]*model.User{&mockUser2}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUsers(2, perPage, "").
			Return([]*model.User{}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Equal(&mockUser1, printer.GetLines()[0])
		s.Require().Equal(&mockUser2, printer.GetLines()[1])
	})

	s.Run("Try to list all the users when there are no uses in store", func() {
		printer.Clean()

		page := 0
		perPage := 1
		showAll := false
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))

		s.client.
			EXPECT().
			GetUsers(page, perPage, "").
			Return([]*model.User{}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Return an error from GetUsers call and verify that error is properly returned", func() {
		printer.Clean()

		page := 0
		perPage := 1
		showAll := false
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))

		mockError := model.AppError{Id: "Mock Error"}
		mockErrorW := errors.Wrap(&mockError, "Failed to fetch users")

		s.client.
			EXPECT().
			GetUsers(page, perPage, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().NotNil(err)
		s.Require().EqualError(err, mockErrorW.Error())
	})

	s.Run("Start with page 2 where a server has total 3 pages", func() {
		printer.Clean()

		email := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: email}

		page := 2
		perPage := 1
		showAll := false
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))

		s.client.
			EXPECT().
			GetUsers(page, perPage, "").
			Return([]*model.User{&mockUser}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
	})

	s.Run("Listing users for given team", func() {
		printer.Clean()

		email := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: email}
		resultID := "teamId"

		page := 0
		perPage := 1
		showAll := false
		team := "teamName"
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))
		_ = cmd.Flags().Set("team", team)

		s.client.
			EXPECT().
			GetTeamByName(team, "").
			Return(&model.Team{Id: resultID}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUsersInTeam(resultID, page, perPage, "").
			Return([]*model.User{&mockUser}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestUserDeactivateCmd() {
	s.Run("Deactivate an existing user using email", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{mockUser.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
	s.Run("Deactivate an existing user by username", func() {
		printer.Clean()
		emailArg := "example@exam.com"
		usernameArg := "ExampleUser"
		mockUser := model.User{Username: usernameArg, Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(usernameArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(usernameArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{mockUser.Username})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Deactivate an existing user by id", func() {
		printer.Clean()
		mockUser := model.User{Id: "userId1", Username: "ExampleUser", Email: "example@exam.com"}

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Id, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(mockUser.Id, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Deactivate SSO user", func() {
		printer.Clean()
		arg := "example@example.com"
		mockUser := model.User{Id: "example-user", Username: "ExampleUser", Email: arg, AuthService: "SSO"}

		s.client.
			EXPECT().
			GetUserByEmail(arg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("You must also deactivate user "+mockUser.Id+" in the SSO provider or they will be reactivated on next login or sync.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Deactivate nonexistent user", func() {
		printer.Clean()
		arg := "example@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* user %s not found\n\n", arg), printer.GetErrorLines()[0])
	})

	s.Run("Delete multiple users", func() {
		printer.Clean()
		mockUser1 := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}
		mockUser2 := model.User{Id: "userId2", Email: "user2@example.com", Username: "user2"}
		mockUser3 := model.User{Id: "userId3", Email: "user3@example.com", Username: "user3"}

		argEmails := []string{mockUser1.Email, mockUser2.Email, mockUser3.Email}
		argUsers := []model.User{mockUser1, mockUser2, mockUser3}

		for i := 0; i < len(argEmails); i++ {
			s.client.
				EXPECT().
				GetUserByEmail(argEmails[i], "").
				Return(&argUsers[i], &model.Response{Error: nil}).
				Times(1)
		}

		for i := 0; i < len(argEmails); i++ {
			s.client.
				EXPECT().
				UpdateUserActive(argUsers[i].Id, false).
				Return(true, &model.Response{Error: nil}).
				Times(1)
		}

		err := userDeactivateCmdF(s.client, &cobra.Command{}, argEmails)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Delete multiple users with argument mixture of emails usernames and userIds", func() {
		printer.Clean()
		mockUser1 := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}
		mockUser2 := model.User{Id: "userId2", Email: "user2@example.com", Username: "user2"}
		mockUser3 := model.User{Id: "userId3", Email: "user3@example.com", Username: "user3"}

		argsDelete := []string{mockUser1.Id, mockUser2.Email, mockUser3.Username}
		argUsers := []model.User{mockUser1, mockUser2, mockUser3}

		// mockUser1
		s.client.
			EXPECT().
			GetUserByEmail(argsDelete[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(argsDelete[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(argsDelete[0], "").
			Return(&argUsers[0], &model.Response{Error: nil}).
			Times(1)

		// mockUser2
		s.client.
			EXPECT().
			GetUserByEmail(argsDelete[1], "").
			Return(&argUsers[1], &model.Response{Error: nil}).
			Times(1)

		// mockUser3
		s.client.
			EXPECT().
			GetUserByEmail(argsDelete[2], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByUsername(argsDelete[2], "").
			Return(&argUsers[2], &model.Response{Error: nil}).
			Times(1)

		for _, user := range argUsers {
			s.client.
				EXPECT().
				UpdateUserActive(user.Id, false).
				Return(true, &model.Response{Error: nil}).
				Times(1)
		}

		err := userDeactivateCmdF(s.client, &cobra.Command{}, argsDelete)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Delete multiple users with an non existent user", func() {
		printer.Clean()
		mockUser1 := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}
		nonexistentEmail := "example@example.com"

		// mockUser1
		s.client.
			EXPECT().
			GetUserByEmail(mockUser1.Email, "").
			Return(&mockUser1, &model.Response{Error: nil}).
			Times(1)

		// nonexistent email
		s.client.
			EXPECT().
			GetUserByEmail(nonexistentEmail, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(nonexistentEmail, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(nonexistentEmail, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser1.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{mockUser1.Email, nonexistentEmail})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* user %s not found\n\n", nonexistentEmail), printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestVerifyUserEmailWithoutTokenCmd() {
	s.Run("Verify user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			VerifyUserEmailWithoutToken(mockUser.Id).
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		err := verifyUserEmailWithoutTokenCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Couldn't find the user", func() {
		printer.Clean()
		userArg := "bad-user-id"

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{StatusCode: http.StatusNotFound}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{StatusCode: http.StatusNotFound}}).
			Times(1)

		err := verifyUserEmailWithoutTokenCmdF(s.client, &cobra.Command{}, []string{userArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
	})

	s.Run("Could not verify user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			VerifyUserEmailWithoutToken(mockUser.Id).
			Return(nil, &model.Response{Error: &model.AppError{Message: "some-message"}}).
			Times(1)

		err := verifyUserEmailWithoutTokenCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Contains(printer.GetErrorLines()[0], fmt.Sprintf("unable to verify user %s email", mockUser.Id))
	})
}

func (s *MmctlUnitTestSuite) TestUserConvertCmd() {
	s.Run("convert user to a bot", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Email: emailArg}
		mockBot := model.Bot{UserId: "example"}

		cmd := &cobra.Command{}
		cmd.Flags().Bool("bot", true, "")

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			ConvertUserToBot(mockUser.Id).
			Return(&mockBot, &model.Response{Error: nil}).
			Times(1)

		err := userConvertCmdF(s.client, cmd, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockBot, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("convert bot to a user", func() {
		printer.Clean()
		userNameArg := "example-bot"
		mockUser := model.User{Id: "example", Email: "example@example.com"}
		mockBot := model.Bot{UserId: "example", Username: userNameArg}
		mockBotUser := model.User{Id: "example", Username: userNameArg, IsBot: true}

		userPatch := model.UserPatch{
			Email:    model.NewString("example@example.com"),
			Password: model.NewString("password"),
			Username: model.NewString("example-user"),
		}

		cmd := &cobra.Command{}
		cmd.Flags().Bool("user", true, "")
		cmd.Flags().String("password", "password", "")
		cmd.Flags().String("email", "example@example.com", "")
		cmd.Flags().String("username", "example-user", "")

		s.client.
			EXPECT().
			GetUserByEmail(userNameArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userNameArg, "").
			Return(&mockBotUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			ConvertBotToUser(mockBot.UserId, &userPatch, false).
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		err := userConvertCmdF(s.client, cmd, []string{userNameArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("fail for not providing either user or bot flag", func() {
		printer.Clean()
		emailArg := "example@example.com"

		cmd := &cobra.Command{}

		err := userConvertCmdF(s.client, cmd, []string{emailArg})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("got error while converting a user to a bot", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Email: emailArg}

		cmd := &cobra.Command{}
		cmd.Flags().Bool("bot", true, "")

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			ConvertUserToBot(mockUser.Id).
			Return(nil, &model.Response{Error: &model.AppError{Message: "some-error"}}).
			Times(1)

		err := userConvertCmdF(s.client, cmd, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
	})

	s.Run("got error while converting a bot to a user", func() {
		printer.Clean()
		userNameArg := "example-bot"
		mockBot := model.Bot{UserId: "example", Username: userNameArg}
		mockBotUser := model.User{Id: "example", Username: userNameArg, IsBot: true}

		userPatch := model.UserPatch{
			Email:    model.NewString("example@example.com"),
			Password: model.NewString("password"),
			Username: model.NewString("example-user"),
		}

		cmd := &cobra.Command{}
		cmd.Flags().Bool("user", true, "")
		cmd.Flags().String("password", "password", "")
		cmd.Flags().String("email", "example@example.com", "")
		cmd.Flags().String("username", "example-user", "")

		s.client.
			EXPECT().
			GetUserByEmail(userNameArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userNameArg, "").
			Return(&mockBotUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			ConvertBotToUser(mockBot.UserId, &userPatch, false).
			Return(nil, &model.Response{Error: &model.AppError{Message: "some-error"}}).
			Times(1)

		err := userConvertCmdF(s.client, cmd, []string{userNameArg})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestMigrateAuthCmd() {
	s.Run("Successfully convert auth to LDAP", func() {
		printer.Clean()

		fromAuth := "email"
		toAuth := "ldap"
		matchField := "username"

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "")

		s.client.
			EXPECT().
			MigrateAuthToLdap(fromAuth, matchField, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := migrateAuthCmdF(s.client, cmd, []string{fromAuth, toAuth, matchField})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Successfully convert auth to SAML", func() {
		printer.Clean()

		fromAuth := "email"
		toAuth := "saml"

		file, err := ioutil.TempFile("", "users.json")
		s.Require().NoError(err)
		defer os.Remove(file.Name())
		usersFile := file.Name()

		userData := map[string]string{
			"usr1@email.com": "usr.one",
			"usr2@email.com": "usr.two",
		}
		b, err := json.MarshalIndent(userData, "", "  ")
		s.Require().NoError(err)

		_, err = file.Write(b)
		s.Require().NoError(err)

		err = file.Sync()
		s.Require().NoError(err)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("auto", false, "")

		s.client.
			EXPECT().
			MigrateAuthToSaml(fromAuth, userData, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err = migrateAuthCmdF(s.client, cmd, []string{fromAuth, toAuth, usersFile})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Successfully convert auth to SAML (auto)", func() {
		printer.Clean()

		fromAuth := "email"
		toAuth := "saml"

		cmd := &cobra.Command{}
		cmd.Flags().Bool("auto", true, "")
		cmd.Flags().Bool("confirm", true, "")

		s.client.
			EXPECT().
			MigrateAuthToSaml(fromAuth, map[string]string{}, true).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := migrateAuthCmdF(s.client, cmd, []string{fromAuth, toAuth})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Invalid from auth type", func() {
		printer.Clean()

		fromAuth := "onelogin"
		toAuth := "ldap"
		matchField := "username"

		cmd := &cobra.Command{}
		cmd.Flags().Bool("auto", true, "")
		cmd.Flags().Bool("confirm", true, "")

		err := migrateAuthCmdF(s.client, cmd, []string{fromAuth, toAuth, matchField})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Invalid matchfiled type for migrating auth to LDAP", func() {
		printer.Clean()

		fromAuth := "email"
		toAuth := "ldap"
		matchField := "groups"

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "")

		err := migrateAuthCmdF(s.client, cmd, []string{fromAuth, toAuth, matchField})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Fail on convert auth to SAML due to invalid file", func() {
		printer.Clean()

		fromAuth := "email"
		toAuth := "saml"
		usersFile := "./nofile.json"

		cmd := &cobra.Command{}
		cmd.Flags().Bool("auto", false, "")

		err := migrateAuthCmdF(s.client, cmd, []string{fromAuth, toAuth, usersFile})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Failed to convert auth to LDAP from server", func() {
		printer.Clean()

		fromAuth := "email"
		toAuth := "ldap"
		matchField := "username"

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "")

		s.client.
			EXPECT().
			MigrateAuthToLdap(fromAuth, matchField, false).
			Return(false, &model.Response{Error: &model.AppError{Id: "some-error"}}).
			Times(1)

		err := migrateAuthCmdF(s.client, cmd, []string{fromAuth, toAuth, matchField})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Failed to convert auth to SAML (auto) from server", func() {
		printer.Clean()

		fromAuth := "email"
		toAuth := "saml"

		cmd := &cobra.Command{}
		cmd.Flags().Bool("auto", true, "")
		cmd.Flags().Bool("confirm", true, "")

		s.client.
			EXPECT().
			MigrateAuthToSaml(fromAuth, map[string]string{}, true).
			Return(false, &model.Response{Error: &model.AppError{Id: "some-error"}}).
			Times(1)

		err := migrateAuthCmdF(s.client, cmd, []string{fromAuth, toAuth})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestPromoteGuestToUserCmd() {
	s.Run("promote a guest to a user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PromoteGuestToUser(mockUser.Id).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := promoteGuestToUserCmdF(s.client, nil, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("cannot promote a guest to a user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Email: emailArg}
		errResp := &model.Response{Error: &model.AppError{Message: "some-error"}}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PromoteGuestToUser(mockUser.Id).
			Return(false, errResp).
			Times(1)

		err := promoteGuestToUserCmdF(s.client, nil, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Sprintf("unable to promote guest %s: %s", emailArg, errResp.Error), printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestDemoteUserToGuestCmd() {
	s.Run("demote a user to a guest", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DemoteUserToGuest(mockUser.Id).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := demoteUserToGuestCmdF(s.client, nil, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("cannot demote a user to a guest", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Email: emailArg}
		errResp := &model.Response{Error: &model.AppError{Message: "some-error"}}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DemoteUserToGuest(mockUser.Id).
			Return(false, errResp).
			Times(1)

		err := demoteUserToGuestCmdF(s.client, nil, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Sprintf("unable to demote user %s: %s", emailArg, errResp.Error), printer.GetErrorLines()[0])
	})
}
