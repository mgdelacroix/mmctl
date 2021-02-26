// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

//nolint:gosec
package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/app"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils"

	"github.com/icrowley/fake"
	"github.com/spf13/cobra"
)

const (
	deactivatedUser = "deactivated"
	guestUser       = "guest"
	attachmentsDir  = "attachments"
)

var SampledataCmd = &cobra.Command{
	Use:   "sampledata",
	Short: "Generate sample data",
	Long:  "Generate a sample data file and store it locally, or directly import it to the remote server",
	Example: `  # you can create a sampledata file and store it locally
  $ mmctl sampledata --bulk sampledata-file.jsonl

  # or you can simply print it to the stdout
  $ mmctl sampledata --bulk -

  # the amount of entities to create can be customized
  $ mmctl sampledata -t 7 -u 20 -g 4

  # the sampledata file can be directly imported in the remote server by not specifying a --bulk flag
  $ mmctl sampledata

  # and the sample users can be created with profile pictures
  $ mmctl sampledata --profile-images`,
	Args: cobra.NoArgs,
	RunE: withClient(sampledataCmdF),
}

func init() {
	SampledataCmd.Flags().Int64P("seed", "s", 1, "Seed used for generating the random data (Different seeds generate different data).")
	SampledataCmd.Flags().IntP("teams", "t", 2, "The number of sample teams.")
	SampledataCmd.Flags().Int("channels-per-team", 10, "The number of sample channels per team.")
	SampledataCmd.Flags().IntP("users", "u", 15, "The number of sample users.")
	SampledataCmd.Flags().IntP("guests", "g", 1, "The number of sample guests.")
	SampledataCmd.Flags().Int("deactivated-users", 0, "The number of deactivated users.")
	SampledataCmd.Flags().Int("team-memberships", 2, "The number of sample team memberships per user.")
	SampledataCmd.Flags().Int("channel-memberships", 5, "The number of sample channel memberships per user in a team.")
	SampledataCmd.Flags().Int("posts-per-channel", 100, "The number of sample post per channel.")
	SampledataCmd.Flags().Int("direct-channels", 30, "The number of sample direct message channels.")
	SampledataCmd.Flags().Int("posts-per-direct-channel", 15, "The number of sample posts per direct message channel.")
	SampledataCmd.Flags().Int("group-channels", 15, "The number of sample group message channels.")
	SampledataCmd.Flags().Int("posts-per-group-channel", 30, "The number of sample posts per group message channel.")
	SampledataCmd.Flags().String("profile-images", "", "Optional. Path to folder with images to randomly pick as user profile image.")
	SampledataCmd.Flags().StringP("bulk", "b", "", "Optional. Path to write a JSONL bulk file instead of uploading into the remote server.")

	RootCmd.AddCommand(SampledataCmd)
}

func randomPastTime(seconds int) int64 {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.FixedZone("UTC", 0))
	return (today.Unix() * 1000) - int64(rand.Intn(seconds*1000))
}

func sortedRandomDates(size int) []int64 {
	dates := make([]int64, size)
	for i := 0; i < size; i++ {
		dates[i] = randomPastTime(50000)
	}
	sort.Slice(dates, func(a, b int) bool { return dates[a] < dates[b] })
	return dates
}

func randomEmoji() string {
	emojis := []string{"+1", "-1", "heart", "blush"}
	return emojis[rand.Intn(len(emojis))]
}

func randomReaction(users []string, parentCreateAt int64) app.ReactionImportData {
	user := users[rand.Intn(len(users))]
	emoji := randomEmoji()
	date := parentCreateAt + int64(rand.Intn(100000))
	return app.ReactionImportData{
		User:      &user,
		EmojiName: &emoji,
		CreateAt:  &date,
	}
}

func randomReply(users []string, parentCreateAt int64) app.ReplyImportData {
	user := users[rand.Intn(len(users))]
	message := randomMessage(users)
	date := parentCreateAt + int64(rand.Intn(100000))
	return app.ReplyImportData{
		User:     &user,
		Message:  &message,
		CreateAt: &date,
	}
}

func randomMessage(users []string) string {
	var message string
	switch rand.Intn(30) {
	case 0:
		mention := users[rand.Intn(len(users))]
		message = "@" + mention + " " + fake.Sentence()
	case 1:
		switch rand.Intn(2) {
		case 0:
			mattermostVideos := []string{"Q4MgnxbpZas", "BFo7E9-Kc_E", "LsMLR-BHsKg", "MRmGDhlMhNA", "mUOPxT7VgWc"}
			message = "https://www.youtube.com/watch?v=" + mattermostVideos[rand.Intn(len(mattermostVideos))]
		case 1:
			mattermostTweets := []string{"943119062334353408", "949370809528832005", "948539688171819009", "939122439115681792", "938061722027425797"}
			message = "https://twitter.com/mattermosthq/status/" + mattermostTweets[rand.Intn(len(mattermostTweets))]
		}
	case 2:
		message = ""
		if rand.Intn(2) == 0 {
			message += fake.Sentence()
		}
		for i := 0; i < rand.Intn(4)+1; i++ {
			message += "\n  * " + fake.Word()
		}
	default:
		if rand.Intn(2) == 0 {
			message = fake.Sentence()
		} else {
			message = fake.Paragraph()
		}
		if rand.Intn(3) == 0 {
			message += "\n" + fake.Sentence()
		}
		if rand.Intn(3) == 0 {
			message += "\n" + fake.Sentence()
		}
		if rand.Intn(3) == 0 {
			message += "\n" + fake.Sentence()
		}
	}
	return message
}

func createUser(idx int, teamMemberships int, channelMemberships int, teamsAndChannels map[string][]string, profileImages []string, userType string) app.LineImportData {
	firstName := fake.FirstName()
	lastName := fake.LastName()
	position := fake.JobTitle()

	username := fmt.Sprintf("%s.%s", strings.ToLower(firstName), strings.ToLower(lastName))
	roles := "system_user"

	var password string
	var email string

	switch userType {
	case guestUser:
		password = fmt.Sprintf("SampleGu@st-%d", idx)
		email = fmt.Sprintf("guest-%d@sample.mattermost.com", idx)
		roles = "system_guest"
		if idx == 0 {
			username = "guest"
			password = "SampleGu@st1"
			email = "guest@sample.mattermost.com"
		}
	case deactivatedUser:
		password = fmt.Sprintf("SampleDe@ctivated-%d", idx)
		email = fmt.Sprintf("deactivated-%d@sample.mattermost.com", idx)
	default:
		password = fmt.Sprintf("SampleUs@r-%d", idx)
		email = fmt.Sprintf("user-%d@sample.mattermost.com", idx)
		if idx == 0 {
			username = "sysadmin"
			password = "Sys@dmin-sample1"
			email = "sysadmin@sample.mattermost.com"
		} else if idx == 1 {
			username = "user-1"
		}

		if idx%5 == 0 {
			roles = "system_admin system_user"
		}
	}

	// The 75% of the users have custom profile image
	var profileImage *string = nil
	if rand.Intn(4) != 0 {
		profileImageSelector := rand.Int()
		if len(profileImages) > 0 {
			profileImage = &profileImages[profileImageSelector%len(profileImages)]
		}
	}

	useMilitaryTime := "false"
	if idx != 0 && rand.Intn(2) == 0 {
		useMilitaryTime = "true"
	}

	collapsePreviews := "false"
	if idx != 0 && rand.Intn(2) == 0 {
		collapsePreviews = "true"
	}

	messageDisplay := "clean"
	if idx != 0 && rand.Intn(2) == 0 {
		messageDisplay = "compact"
	}

	channelDisplayMode := "full"
	if idx != 0 && rand.Intn(2) == 0 {
		channelDisplayMode = "centered"
	}

	// Some users has nickname
	nickname := ""
	if rand.Intn(5) == 0 {
		nickname = fake.Company()
	}

	// sysadmin, user-1 and user-2 users skip tutorial steps
	// Other half of users also skip tutorial steps
	tutorialStep := "999"
	if idx > 2 {
		switch rand.Intn(6) {
		case 1:
			tutorialStep = "1"
		case 2:
			tutorialStep = "2"
		case 3:
			tutorialStep = "3"
		}
	}

	teams := []app.UserTeamImportData{}
	possibleTeams := []string{}
	for teamName := range teamsAndChannels {
		possibleTeams = append(possibleTeams, teamName)
	}
	sort.Strings(possibleTeams)
	for x := 0; x < teamMemberships; x++ {
		if len(possibleTeams) == 0 {
			break
		}
		position := rand.Intn(len(possibleTeams))
		team := possibleTeams[position]
		possibleTeams = append(possibleTeams[:position], possibleTeams[position+1:]...)
		if teamChannels, err := teamsAndChannels[team]; err {
			teams = append(teams, createTeamMembership(channelMemberships, teamChannels, &team, userType == guestUser))
		}
	}

	var deleteAt int64
	if userType == deactivatedUser {
		deleteAt = model.GetMillis()
	}

	user := app.UserImportData{
		ProfileImage:       profileImage,
		Username:           &username,
		Email:              &email,
		Password:           &password,
		Nickname:           &nickname,
		FirstName:          &firstName,
		LastName:           &lastName,
		Position:           &position,
		Roles:              &roles,
		Teams:              &teams,
		UseMilitaryTime:    &useMilitaryTime,
		CollapsePreviews:   &collapsePreviews,
		MessageDisplay:     &messageDisplay,
		ChannelDisplayMode: &channelDisplayMode,
		TutorialStep:       &tutorialStep,
		DeleteAt:           &deleteAt,
	}
	return app.LineImportData{
		Type: "user",
		User: &user,
	}
}

func createTeamMembership(numOfchannels int, teamChannels []string, teamName *string, guest bool) app.UserTeamImportData {
	roles := "team_user"
	if guest {
		roles = "team_guest"
	} else if rand.Intn(5) == 0 {
		roles = "team_user team_admin"
	}
	channels := []app.UserChannelImportData{}
	teamChannelsCopy := append([]string(nil), teamChannels...)
	for x := 0; x < numOfchannels; x++ {
		if len(teamChannelsCopy) == 0 {
			break
		}
		position := rand.Intn(len(teamChannelsCopy))
		channelName := teamChannelsCopy[position]
		teamChannelsCopy = append(teamChannelsCopy[:position], teamChannelsCopy[position+1:]...)
		channels = append(channels, createChannelMembership(channelName, guest))
	}

	return app.UserTeamImportData{
		Name:     teamName,
		Roles:    &roles,
		Channels: &channels,
	}
}

func createChannelMembership(channelName string, guest bool) app.UserChannelImportData {
	roles := "channel_user"
	if guest {
		roles = "channel_guest"
	} else if rand.Intn(5) == 0 {
		roles = "channel_user channel_admin"
	}
	favorite := rand.Intn(5) == 0

	return app.UserChannelImportData{
		Name:     &channelName,
		Roles:    &roles,
		Favorite: &favorite,
	}
}

func getSampleTeamName(idx int) string {
	for {
		name := fmt.Sprintf("%s-%d", fake.Word(), idx)
		if !model.IsReservedTeamName(name) {
			return name
		}
	}
}

func createTeam(idx int) app.LineImportData {
	displayName := fake.Word()
	name := getSampleTeamName(idx)
	allowOpenInvite := rand.Intn(2) == 0

	description := fake.Paragraph()
	if len(description) > 255 {
		description = description[0:255]
	}

	teamType := "O"
	if rand.Intn(2) == 0 {
		teamType = "I"
	}

	team := app.TeamImportData{
		DisplayName:     &displayName,
		Name:            &name,
		AllowOpenInvite: &allowOpenInvite,
		Description:     &description,
		Type:            &teamType,
	}
	return app.LineImportData{
		Type: "team",
		Team: &team,
	}
}

func createChannel(idx int, teamName string) app.LineImportData {
	displayName := fake.Word()
	name := fmt.Sprintf("%s-%d", fake.Word(), idx)
	header := fake.Paragraph()
	purpose := fake.Paragraph()

	if len(purpose) > 250 {
		purpose = purpose[0:250]
	}

	channelType := "P"
	if rand.Intn(2) == 0 {
		channelType = "O"
	}

	channel := app.ChannelImportData{
		Team:        &teamName,
		Name:        &name,
		DisplayName: &displayName,
		Type:        &channelType,
		Header:      &header,
		Purpose:     &purpose,
	}
	return app.LineImportData{
		Type:    "channel",
		Channel: &channel,
	}
}

func createPost(team string, channel string, allUsers []string, createAt int64) app.LineImportData {
	message := randomMessage(allUsers)
	user := allUsers[rand.Intn(len(allUsers))]

	// Some messages are flagged by an user
	flaggedBy := []string{}
	if rand.Intn(10) == 0 {
		flaggedBy = append(flaggedBy, allUsers[rand.Intn(len(allUsers))])
	}

	reactions := []app.ReactionImportData{}
	if rand.Intn(10) == 0 {
		for {
			reactions = append(reactions, randomReaction(allUsers, createAt))
			if rand.Intn(3) == 0 {
				break
			}
		}
	}

	replies := []app.ReplyImportData{}
	if rand.Intn(10) == 0 {
		for {
			replies = append(replies, randomReply(allUsers, createAt))
			if rand.Intn(4) == 0 {
				break
			}
		}
	}

	post := app.PostImportData{
		Team:      &team,
		Channel:   &channel,
		User:      &user,
		Message:   &message,
		CreateAt:  &createAt,
		FlaggedBy: &flaggedBy,
		Reactions: &reactions,
		Replies:   &replies,
	}
	return app.LineImportData{
		Type: "post",
		Post: &post,
	}
}

func createDirectChannel(members []string) app.LineImportData {
	header := fake.Sentence()

	channel := app.DirectChannelImportData{
		Members: &members,
		Header:  &header,
	}
	return app.LineImportData{
		Type:          "direct_channel",
		DirectChannel: &channel,
	}
}

func createDirectPost(members []string, createAt int64) app.LineImportData {
	message := randomMessage(members)
	user := members[rand.Intn(len(members))]

	// Some messages are flagged by an user
	flaggedBy := []string{}
	if rand.Intn(10) == 0 {
		flaggedBy = append(flaggedBy, members[rand.Intn(len(members))])
	}

	reactions := []app.ReactionImportData{}
	if rand.Intn(10) == 0 {
		for {
			reactions = append(reactions, randomReaction(members, createAt))
			if rand.Intn(3) == 0 {
				break
			}
		}
	}

	replies := []app.ReplyImportData{}
	if rand.Intn(10) == 0 {
		for {
			replies = append(replies, randomReply(members, createAt))
			if rand.Intn(4) == 0 {
				break
			}
		}
	}

	post := app.DirectPostImportData{
		ChannelMembers: &members,
		User:           &user,
		Message:        &message,
		CreateAt:       &createAt,
		FlaggedBy:      &flaggedBy,
		Reactions:      &reactions,
		Replies:        &replies,
	}
	return app.LineImportData{
		Type:       "direct_post",
		DirectPost: &post,
	}
}

func copyFile(src, dst string) error {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, b, 0600)
}

func uploadAndProcess(c client.Client, zipPath string, isLocal bool) error {
	zipFile, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("cannot open import file %q: %w", zipPath, err)
	}
	defer zipFile.Close()

	info, err := zipFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat import file: %w", err)
	}

	userID := "me"
	if isLocal {
		userID = model.UploadNoUserID
	}

	// create session
	us, resp := c.CreateUpload(&model.UploadSession{
		Filename: info.Name(),
		FileSize: info.Size(),
		Type:     model.UploadTypeImport,
		UserId:   userID,
	})
	if resp.Error != nil {
		return fmt.Errorf("failed to create upload session: %w", resp.Error)
	}

	printer.PrintT("Upload session successfully created, ID: {{.Id}} ", us)

	// upload file
	finfo, resp := c.UploadData(us.Id, zipFile)
	if resp.Error != nil {
		return fmt.Errorf("failed to upload data: %w", resp.Error)
	}

	printer.PrintT("Import file successfully uploaded, name: {{.Name}}", finfo)

	// process
	job, resp := c.CreateJob(&model.Job{
		Type: model.JOB_TYPE_IMPORT_PROCESS,
		Data: map[string]string{
			"import_file": us.Id + "_" + finfo.Name,
		},
	})
	if resp.Error != nil {
		return fmt.Errorf("failed to create import process job: %w", resp.Error)
	}

	printer.PrintT("Import process job successfully created, ID: {{.Id}}", job)

	for {
		job, resp = c.GetJob(job.Id)
		if resp.Error != nil {
			return fmt.Errorf("failed to get import job status: %w", resp.Error)
		}

		if job.Status != model.JOB_STATUS_PENDING && job.Status != model.JOB_STATUS_IN_PROGRESS {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if job.Status != model.JOB_STATUS_SUCCESS {
		return fmt.Errorf("job reported non-success status: %s", job.Status)
	}

	printer.PrintT("Sampledata successfully processed", job)

	return nil
}

//nolint:gocyclo
func sampledataCmdF(c client.Client, command *cobra.Command, args []string) error {
	seed, _ := command.Flags().GetInt64("seed")
	bulk, _ := command.Flags().GetString("bulk")
	teams, _ := command.Flags().GetInt("teams")
	channelsPerTeam, _ := command.Flags().GetInt("channels-per-team")
	users, _ := command.Flags().GetInt("users")
	deactivatedUsers, _ := command.Flags().GetInt("deactivated-users")
	guests, _ := command.Flags().GetInt("guests")
	teamMemberships, _ := command.Flags().GetInt("team-memberships")
	channelMemberships, _ := command.Flags().GetInt("channel-memberships")
	postsPerChannel, _ := command.Flags().GetInt("posts-per-channel")
	directChannels, _ := command.Flags().GetInt("direct-channels")
	postsPerDirectChannel, _ := command.Flags().GetInt("posts-per-direct-channel")
	groupChannels, _ := command.Flags().GetInt("group-channels")
	postsPerGroupChannel, _ := command.Flags().GetInt("posts-per-group-channel")
	profileImagesPath, _ := command.Flags().GetString("profile-images")
	withAttachments := profileImagesPath != ""

	if teamMemberships > teams {
		return fmt.Errorf("you can't have more team memberships than teams")
	}
	if channelMemberships > channelsPerTeam {
		return fmt.Errorf("you can't have more channel memberships than channels per team")
	}

	if users < 6 && groupChannels > 0 {
		return fmt.Errorf("you can't have group channels generation with less than 6 users. Use --group-channels 0 or increase the number of users")
	}

	var bulkFile *os.File
	var tmpDir string
	var err error
	switch bulk {
	case "":
		tmpDir, err = ioutil.TempDir("", "mmctl-sampledata-")
		if err != nil {
			return fmt.Errorf("unable to create temporary directory")
		}
		defer os.RemoveAll(tmpDir)

		if withAttachments {
			if err = os.Mkdir(filepath.Join(tmpDir, attachmentsDir), 0755); err != nil {
				return fmt.Errorf("cannot create attachments directory: %w", err)
			}
		}

		bulkFile, err = os.Create(filepath.Join(tmpDir, "import.jsonl"))
		if err != nil {
			return fmt.Errorf("unable to open temporary file: %w", err)
		}
	case "-":
		bulkFile = os.Stdout
	default:
		bulkFile, err = os.OpenFile(bulk, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return fmt.Errorf("unable to write into the %q file: %w", bulk, err)
		}
	}

	profileImages := []string{}
	if profileImagesPath != "" {
		var profileImagesStat os.FileInfo
		profileImagesStat, err = os.Stat(profileImagesPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("profile images folder doesn't exists")
		}
		if !profileImagesStat.IsDir() {
			return fmt.Errorf("profile-images parameters must be a folder filepath")
		}
		var profileImagesFiles []os.FileInfo
		profileImagesFiles, err = ioutil.ReadDir(profileImagesPath)
		if err != nil {
			return fmt.Errorf("invalid profile-images parameter")
		}

		// we need to copy the images to be part of the import zip
		if bulk == "" {
			for _, profileImage := range profileImagesFiles {
				profileImageSrc := filepath.Join(profileImagesPath, profileImage.Name())
				profileImagePath := filepath.Join(attachmentsDir, profileImage.Name())
				profileImageDst := filepath.Join(tmpDir, profileImagePath)
				if err := copyFile(profileImageSrc, profileImageDst); err != nil {
					return fmt.Errorf("cannot copy file %q to %q: %w", profileImageSrc, profileImageDst, err)
				}
				// the path we use in the profile info is relative to the zipfile base
				profileImages = append(profileImages, profileImagePath)
			}
			// we're not importing the resulting file, so we keep the
			// image paths corresponding to the value of the flag
		} else {
			for _, profileImage := range profileImagesFiles {
				profileImages = append(profileImages, filepath.Join(profileImagesPath, profileImage.Name()))
			}
		}
		sort.Strings(profileImages)
	}

	encoder := json.NewEncoder(bulkFile)
	version := 1
	if err := encoder.Encode(app.LineImportData{Type: "version", Version: &version}); err != nil {
		return fmt.Errorf("could not encode version line: %w", err)
	}

	fake.Seed(seed)
	rand.Seed(seed)

	teamsAndChannels := make(map[string][]string)
	for i := 0; i < teams; i++ {
		teamLine := createTeam(i)
		teamsAndChannels[*teamLine.Team.Name] = []string{}
		if err := encoder.Encode(teamLine); err != nil {
			return fmt.Errorf("could not encode team line: %w", err)
		}
	}

	teamsList := []string{}
	for teamName := range teamsAndChannels {
		teamsList = append(teamsList, teamName)
	}
	sort.Strings(teamsList)

	for _, teamName := range teamsList {
		for i := 0; i < channelsPerTeam; i++ {
			channelLine := createChannel(i, teamName)
			teamsAndChannels[teamName] = append(teamsAndChannels[teamName], *channelLine.Channel.Name)
			if err := encoder.Encode(channelLine); err != nil {
				return fmt.Errorf("could not encode channel line: %w", err)
			}
		}
	}

	allUsers := []string{}
	for i := 0; i < users; i++ {
		userLine := createUser(i, teamMemberships, channelMemberships, teamsAndChannels, profileImages, "")
		if err := encoder.Encode(userLine); err != nil {
			return fmt.Errorf("cannot encode user line: %w", err)
		}
		allUsers = append(allUsers, *userLine.User.Username)
	}
	for i := 0; i < guests; i++ {
		userLine := createUser(i, teamMemberships, channelMemberships, teamsAndChannels, profileImages, guestUser)
		if err := encoder.Encode(userLine); err != nil {
			return fmt.Errorf("cannot encode user line: %w", err)
		}
		allUsers = append(allUsers, *userLine.User.Username)
	}
	for i := 0; i < deactivatedUsers; i++ {
		userLine := createUser(i, teamMemberships, channelMemberships, teamsAndChannels, profileImages, deactivatedUser)
		if err := encoder.Encode(userLine); err != nil {
			return fmt.Errorf("cannot encode user line: %w", err)
		}
		allUsers = append(allUsers, *userLine.User.Username)
	}

	for team, channels := range teamsAndChannels {
		for _, channel := range channels {
			dates := sortedRandomDates(postsPerChannel)

			for i := 0; i < postsPerChannel; i++ {
				postLine := createPost(team, channel, allUsers, dates[i])
				if err := encoder.Encode(postLine); err != nil {
					return fmt.Errorf("cannot encode post line: %w", err)
				}
			}
		}
	}

	for i := 0; i < directChannels; i++ {
		user1 := allUsers[rand.Intn(len(allUsers))]
		user2 := allUsers[rand.Intn(len(allUsers))]
		channelLine := createDirectChannel([]string{user1, user2})
		if err := encoder.Encode(channelLine); err != nil {
			return fmt.Errorf("cannot encode channel line: %w", err)
		}
	}

	for i := 0; i < directChannels; i++ {
		user1 := allUsers[rand.Intn(len(allUsers))]
		user2 := allUsers[rand.Intn(len(allUsers))]

		dates := sortedRandomDates(postsPerDirectChannel)
		for j := 0; j < postsPerDirectChannel; j++ {
			postLine := createDirectPost([]string{user1, user2}, dates[j])
			if err := encoder.Encode(postLine); err != nil {
				return fmt.Errorf("cannot encode post line: %w", err)
			}
		}
	}

	for i := 0; i < groupChannels; i++ {
		users := []string{}
		totalUsers := 3 + rand.Intn(3)
		for len(users) < totalUsers {
			user := allUsers[rand.Intn(len(allUsers))]
			if !utils.StringInSlice(user, users) {
				users = append(users, user)
			}
		}
		channelLine := createDirectChannel(users)
		if err := encoder.Encode(channelLine); err != nil {
			return fmt.Errorf("cannot encode channel line: %w", err)
		}
	}

	for i := 0; i < groupChannels; i++ {
		users := []string{}
		totalUsers := 3 + rand.Intn(3)
		for len(users) < totalUsers {
			user := allUsers[rand.Intn(len(allUsers))]
			if !utils.StringInSlice(user, users) {
				users = append(users, user)
			}
		}

		dates := sortedRandomDates(postsPerGroupChannel)
		for j := 0; j < postsPerGroupChannel; j++ {
			postLine := createDirectPost(users, dates[j])
			if err := encoder.Encode(postLine); err != nil {
				return fmt.Errorf("cannot encode post line: %w", err)
			}
		}
	}

	// if we're writing to stdout, we can finish here
	if bulk == "-" {
		return nil
	}

	if bulk == "" {
		zipPath := filepath.Join(os.TempDir(), "mmctl-sampledata.zip")
		defer os.Remove(zipPath)

		if err := zipDir(zipPath, tmpDir); err != nil {
			return fmt.Errorf("cannot compress %q directory into zipfile: %w", tmpDir, err)
		}

		isLocal, _ := command.Flags().GetBool("local")
		if err := uploadAndProcess(c, zipPath, isLocal); err != nil {
			return fmt.Errorf("cannot upload and process zipfile: %w", err)
		}
	} else if bulk != "-" {
		err := bulkFile.Close()
		if err != nil {
			return fmt.Errorf("unable to close correctly the output file")
		}
	}

	return nil
}
