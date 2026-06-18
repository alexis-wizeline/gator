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
	loginCommandName     = "login"
	registerCommandName  = "register"
	resetCommandName     = "reset"
	usersCommandName     = "users"
	aggCommandName       = "agg"
	addfeedCommandName   = "addfeed"
	feedsCommandName     = "feeds"
	followCommandName    = "follow"
	followingCommandName = "following"
	unfollowCommandName  = "unfollow"
	browseCommandName    = "browse"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Unable to read config file: %s\n", err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		fmt.Printf("unable to connect to: %s\n", cfg.DBUrl)
		os.Exit(1)
	}

	dbQueries := gatordb.New(db)
	state := state.NewState(dbQueries, cfg)
	gator := commands.GatorCommands()

	gator.Register(loginCommandName, commands.HandlerLogin)
	gator.Register(registerCommandName, commands.HandlerRegister)
	gator.Register(resetCommandName, commands.HandleReset)
	gator.Register(usersCommandName, commands.HandleUsers)
	gator.Register(aggCommandName, commands.HandleAgg)
	gator.Register(addfeedCommandName, commands.GetUserMiddleware(commands.HandleAddFeed))
	gator.Register(feedsCommandName, commands.HandleFeeds)
	gator.Register(followCommandName, commands.GetUserMiddleware(commands.HandleFollow))
	gator.Register(followingCommandName, commands.GetUserMiddleware(commands.HandleFollowing))
	gator.Register(unfollowCommandName, commands.GetUserMiddleware(commands.HandleUnfollow))
	gator.Register(browseCommandName, commands.GetUserMiddleware(commands.HandleBrowse))

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
