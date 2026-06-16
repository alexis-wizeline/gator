package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/alexis-wizeline/gator/internal/commands"
	"github.com/alexis-wizeline/gator/internal/config"
	"github.com/alexis-wizeline/gator/internal/gatordb"
	"github.com/alexis-wizeline/gator/internal/state"
)

const (
	loginCommandName    = "login"
	registerCommandName = "register"
	resetCommandName    = "reset"
	usersCommandName    = "users"
	aggCommandName      = "agg"
	addfeedCommandName  = "addfeed"
	feedsCommandName    = "feeds"
)

func main() {
	config := config.Read()
	db, err := sql.Open("postgres", config.DBUrl)
	if err != nil {
		fmt.Printf("unable to connect to: %s\n", config.DBUrl)
		os.Exit(1)
	}

	dbQueries := gatordb.New(db)
	state := state.NewState(dbQueries, config)
	gator := commands.GatorCommands()

	gator.Register(loginCommandName, commands.HandlerLogin)
	gator.Register(registerCommandName, commands.HandlerRegister)
	gator.Register(resetCommandName, commands.HandleReset)
	gator.Register(usersCommandName, commands.HandleUsers)
	gator.Register(aggCommandName, commands.HandleAgg)
	gator.Register(addfeedCommandName, commands.HandleAddFeed)
	gator.Register(feedsCommandName, commands.HandleFeeds)

	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("a command is required")
		os.Exit(1)
	}

	ctx := context.Background()
	inputCommand := args[0]
	err = gator.Run(ctx, state, commands.Command{
		Name:      inputCommand,
		Arguments: args[1:],
	})
	if err != nil {
		fmt.Printf("%s command failed with: %s\n", inputCommand, err)
		os.Exit(1)
	}
}
