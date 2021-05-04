// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
package commands

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlUnitTestSuite) TestGetUserFromArgs() {
	s.Run("user not found", func() {
		notFoundEmail := "emailNotfound@notfound.com"
		notFoundErr := &model.AppError{Message: "user not found", StatusCode: http.StatusNotFound}
		printer.Clean()
		s.client.
			EXPECT().
			GetUserByEmail(notFoundEmail, "").
			Return(nil, &model.Response{Error: notFoundErr}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByUsername(notFoundEmail, "").
			Return(nil, &model.Response{Error: notFoundErr}).
			Times(1)
		s.client.
			EXPECT().
			GetUser(notFoundEmail, "").
			Return(nil, &model.Response{Error: notFoundErr}).
			Times(1)

		users, err := getUsersFromArgs(s.client, []string{notFoundEmail})
		s.Require().Empty(users)
		s.Require().NotNil(err)
		s.Require().EqualError(err, fmt.Sprintf("1 error occurred:\n\t* user %s not found\n\n", notFoundEmail))
	})

	s.Run("bad request don't throw unexpected error", func() {
		badRequestEmail := "emailbadrequest@badrequest.com"
		badRequestErr := &model.AppError{Message: "bad request", StatusCode: http.StatusBadRequest}
		printer.Clean()
		s.client.
			EXPECT().
			GetUserByEmail(badRequestEmail, "").
			Return(nil, &model.Response{Error: badRequestErr}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByUsername(badRequestEmail, "").
			Return(nil, &model.Response{Error: badRequestErr}).
			Times(1)
		s.client.
			EXPECT().
			GetUser(badRequestEmail, "").
			Return(nil, &model.Response{Error: badRequestErr}).
			Times(1)

		users, err := getUsersFromArgs(s.client, []string{badRequestEmail})
		s.Require().Empty(users)
		s.Require().NotNil(err)
		s.Require().EqualError(err, fmt.Sprintf("1 error occurred:\n\t* user %s not found\n\n", badRequestEmail))
	})

	s.Run("unexpected error throws according error", func() {
		unexpectedErrEmail := "emailunexpected@unexpected.com"
		unexpectedErr := &model.AppError{Message: "internal server error", StatusCode: http.StatusInternalServerError}
		printer.Clean()
		s.client.
			EXPECT().
			GetUserByEmail(unexpectedErrEmail, "").
			Return(nil, &model.Response{Error: unexpectedErr}).
			Times(1)
		users, err := getUsersFromArgs(s.client, []string{unexpectedErrEmail})
		s.Require().Empty(users)
		s.Require().NotNil(err)
		s.Require().EqualError(err, "1 error occurred:\n\t* : internal server error, \n\n")
	})
	s.Run("forbidden error stops searching", func() {
		forbiddenErrEmail := "forbidden@forbidden.com"
		forbiddenErr := &model.AppError{Message: "forbidden", StatusCode: http.StatusForbidden}
		printer.Clean()
		s.client.
			EXPECT().
			GetUserByEmail(forbiddenErrEmail, "").
			Return(nil, &model.Response{Error: forbiddenErr}).
			Times(1)
		users, err := getUsersFromArgs(s.client, []string{forbiddenErrEmail})
		s.Require().Empty(users)
		s.Require().NotNil(err)
		s.Require().EqualError(err, "1 error occurred:\n\t* : forbidden, \n\n")
	})
	s.Run("success", func() {
		successEmail := "success@success.com"
		successUser := &model.User{Email: successEmail}
		printer.Clean()
		s.client.
			EXPECT().
			GetUserByEmail(successEmail, "").
			Return(successUser, nil).
			Times(1)
		users, err := getUsersFromArgs(s.client, []string{successEmail})
		s.Require().NoError(err)
		s.Require().Len(users, 1)
		s.Require().Equal(successUser, users[0])
	})
}
