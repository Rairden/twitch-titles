package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/nicklaw5/helix"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
)

var (
	oauth2Config *clientcredentials.Config
	clientID     string
	clientSecret string
)

const (
	basicProgrammingID           = "21548"
	pubgID                       = "493057"
	scienceAndTechnologyID       = "509670"
	softwareAndGameDevelopmentID = "1469308723"
)

func main() {
	clientID = os.Getenv("TWITCH_CLIENTID")
	clientSecret = os.Getenv("TWITCH_CLIENTSECRET")
	client := InitTwitchClient()
	args := os.Args[1:]

	regex := ".*"
	var games []string
	if len(args) == 1 {
		games = strings.Split(args[0], ",")
	} else {
		regex = args[0]
		games = strings.Split(args[1], ",")
	}

	for i := range games {
		games[i] = strings.TrimSpace(games[i])
	}

	manyGames, err := findGame(client, games)
	if err != nil {
		fmt.Println("Could not find any games")
		os.Exit(1)
	}

	var categories []string
	for _, game := range manyGames {
		categories = append(categories, game.ID)
	}

	findStreams(client, categories, regex)
}

func findStreams(client *helix.Client, categories []string, regex string) {
	search := &helix.StreamsParams{GameIDs: categories}
	result, _ := client.GetStreams(search)
	currCursor := result.Data.Pagination.Cursor
	re := regexp.MustCompile(regex)

	// A hack. The cursor is an empty string if it's only 1 stream
	if len(result.Data.Streams) == 1 {
		currCursor = "null"
	}

	matches, processed := 0, 0
	for currCursor != "" {
		for _, s := range result.Data.Streams {
			processed++
			if re.MatchString(s.Title) {
				matches++
				stream := fmt.Sprintf("%4d https://www.twitch.tv/%-25s %s", s.ViewerCount, s.UserLogin, s.Title)
				fmt.Println(stream)
			}
		}
		search = &helix.StreamsParams{
			GameIDs: categories,
			After:   currCursor,
		}

		result, _ = client.GetStreams(search)
		currCursor = result.Data.Pagination.Cursor
	}

	streams := "streams"
	results := "results"
	if matches == 1 {
		results = "result"
	}
	if processed == 1 {
		streams = "stream"
	}

	if regex != ".*" {
		msg := fmt.Sprintf("\n%d %s in %d %s.", matches, results, processed, streams)
		fmt.Println(msg)
	} else {
		msg := fmt.Sprintf("\n%d %s.", processed, results)
		fmt.Println(msg)
	}
}

func findGame(client *helix.Client, categories []string) ([]helix.Game, error) {
	search := &helix.GamesParams{Names: categories}
	result, _ := client.GetGames(search)
	if len(result.Data.Games) == 0 {
		return []helix.Game{}, errors.New("error finding game")
	}
	return result.Data.Games, nil
}

func findTopGames(client *helix.Client) {
	param := &helix.TopGamesParams{First: 100}

	games, _ := client.GetTopGames(param)
	manyGames := games.Data.Games
	for _, game := range manyGames {
		fmt.Println(game.ID, game.Name)
	}
}

func searchChannels(client *helix.Client) {
	resp, err := client.SearchChannels(&helix.SearchChannelsParams{
		First: 2,
	})
	if err != nil {
		// handle error
	}

	fmt.Printf("%+v\n", resp)
}

func findUsers(client *helix.Client) {
	userResponse, userResponseError := client.GetUsers(&helix.UsersParams{
		Logins: []string{"summit1g", "rairden"},
	})

	if userResponseError != nil {
		fmt.Printf("Error getting users: %s", userResponseError)
	}

	for _, user := range userResponse.Data.Users {
		fmt.Printf("DisplayName: %s Id: %s\n", user.DisplayName, user.ID)
	}
}

func InitTwitchClient() *helix.Client {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})

	if err != nil {
		panic(err)
	}

	tokenResponse, tokenResponseErr := client.RequestAppAccessToken([]string{})
	if tokenResponseErr != nil || tokenResponse.StatusCode != 200 {
		panic(tokenResponseErr)
	}

	client.SetAppAccessToken(tokenResponse.Data.AccessToken)
	return client
}

func getToken() {
	oauth2Config = &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     twitch.Endpoint.TokenURL,
	}

	token, err := oauth2Config.Token(context.Background())
	if err != nil {
		fmt.Println("token error")
	}
	fmt.Printf("Access token: %s\n", token.AccessToken)
}
