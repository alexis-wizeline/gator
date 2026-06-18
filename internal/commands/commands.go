package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/alexis-wizeline/gator/internal/gatordb"
	"github.com/alexis-wizeline/gator/internal/rss"
	"github.com/alexis-wizeline/gator/internal/state"
)

type handler struct {
	f func(context.Context, *state.State, Command) error
}

type Commands struct {
	handlers map[string]handler
}

func (c *Commands) Run(ctx context.Context, s *state.State, cmd Command) error {
	handler, ok := c.handlers[cmd.Name]
	if !ok {
		return errors.New("unknow command")
	}
	return handler.f(ctx, s, cmd)
}

func (c *Commands) Register(name string, f func(context.Context, *state.State, Command) error) {
	c.handlers[name] = handler{f}
}

func GatorCommands() *Commands {
	handlers := make(map[string]handler)
	return &Commands{
		handlers,
	}
}

type Command struct {
	Name      string
	Arguments []string
	User      gatordb.User
}

func HandlerLogin(ctx context.Context, s *state.State, cmd Command) error {
	if len(cmd.Arguments) < 1 {
		return errors.New("the parameter username is required")
	}

	if len(cmd.Arguments) > 1 {
		return errors.New("this command only require username as parameter")
	}

	username := cmd.Arguments[0]
	exist, err := checkUserExist(ctx, s, username)
	if !exist {
		return fmt.Errorf("user %s is not yet register", username)
	}
	if err != nil {
		return fmt.Errorf("GetUserByName Failed: %w", err)
	}

	err = s.Config.SetUser(username)
	if err != nil {
		return fmt.Errorf("Config.SetUser Failed: %w", err)
	}

	fmt.Printf("welcome %s you are login now..\n", cmd.Arguments[0])
	return nil
}

func HandlerRegister(ctx context.Context, s *state.State, cmd Command) error {
	if len(cmd.Arguments) < 1 {
		return errors.New("the users to register are required.")
	}

	name := cmd.Arguments[0]
	exists, err := checkUserExist(ctx, s, name)
	if exists {
		return fmt.Errorf("user %s already exists", name)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("GetUserByName Failed: %w", err)
	}

	currentTime := time.Now()
	u, err := s.DB.CreateUser(ctx, gatordb.CreateUserParams{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return fmt.Errorf("CreateUser failed: %w", err)
	}

	fmt.Printf(`
		User: %s
		was created with ID: %s

		`, u.Name, u.ID)

	return s.Config.SetUser(name)
}

func HandleReset(ctx context.Context, s *state.State, _ Command) error {
	return s.DB.DeleteUsers(ctx)
}

func HandleUsers(ctx context.Context, s *state.State, _ Command) error {
	users, err := s.DB.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("GetUsers Failed: %w", err)
	}

	for _, user := range users {
		if s.Config.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %s\n", user.Name)
	}
	return nil
}

func HandleAgg(ctx context.Context, s *state.State, c Command) error {
	if len(c.Arguments) < 1 {
		return errors.New("the interval is required")
	}

	duration, err := time.ParseDuration(c.Arguments[0])
	if err != nil {
		return fmt.Errorf("unable to parse the provided duration %w", err)
	}

	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		err = scrapeFeeds(ctx, s)
		if err != nil {
			return fmt.Errorf("scrapeFeeds Failed: %w", err)
		}
	}
}

func HandleAddFeed(ctx context.Context, s *state.State, c Command) error {
	if len(c.Arguments) < 2 {
		return errors.New("the addfeed command must have 2 argumenst name and url")
	}
	name := c.Arguments[0]
	url := c.Arguments[1]
	feed, err := s.DB.CreateFeed(ctx, gatordb.CreateFeedParams{
		ID:     uuid.New(),
		Name:   name,
		Url:    url,
		UserID: c.User.ID,
	})
	if err != nil {
		return fmt.Errorf("CreateFedd Failed: %w", err)
	}

	_, err = s.DB.CreateFeedFollow(ctx,
		gatordb.CreateFeedFollowParams{
			UserID: c.User.ID,
			FeedID: feed.ID,
		})
	if err != nil {
		return fmt.Errorf("CreateFeedFollow Failed: %w", err)
	}

	fmt.Printf("feed added: %v\n", feed)

	return nil
}

func HandleFeeds(ctx context.Context, s *state.State, _ Command) error {
	feeds, err := s.DB.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("GetFeeds Failed: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("%s - %s, Crated by: %s\n", feed.Name, feed.Url, feed.User)
	}

	return nil
}

func HandleFollow(ctx context.Context, s *state.State, c Command) error {
	if len(c.Arguments) < 1 {
		return errors.New("the feed url to follow is required")
	}

	feed, err := s.DB.GetFeedByURL(ctx, c.Arguments[0])
	if err != nil {
		return fmt.Errorf("GetFeedByURL Failed: %w", err)
	}

	follow, err := s.DB.CreateFeedFollow(ctx,
		gatordb.CreateFeedFollowParams{
			UserID: c.User.ID,
			FeedID: feed.ID,
		})
	if err != nil {
		return fmt.Errorf("CreateFeedFollow Failed: %w", err)
	}

	fmt.Printf("user: %s, now follows the Feed: %s\n", follow.Username, follow.FeedName)

	return nil
}

func HandleFollowing(ctx context.Context, s *state.State, c Command) error {

	follows, err := s.DB.GetFeedFollowsByUser(ctx, c.User.ID)
	if err != nil {
		return fmt.Errorf("GetFeedFollowsByUser Failed: %w", err)
	}

	for _, follow := range follows {
		fmt.Printf("%s\n", follow.Name)
	}

	return nil
}

func HandleUnfollow(ctx context.Context, s *state.State, c Command) error {
	if len(c.Arguments) < 1 {
		return errors.New("the feed url is needed")
	}

	feed, err := s.DB.GetFeedByURL(ctx, c.Arguments[0])
	if err != nil {
		return fmt.Errorf("GetFeedByURL Failed: %w", err)
	}

	err = s.DB.DeleteFeedFollow(ctx, gatordb.DeleteFeedFollowParams{
		UserID: c.User.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("DeleteFeedFollow Failed: %w", err)
	}

	return nil
}

func HandleBrowser(ctx context.Context, s *state.State, c Command) error {
	if len(c.Arguments) < 1 {
		return errors.New("The limit number si required")
	}

	limit, err := strconv.Atoi(c.Arguments[0])
	if err != nil {
		return fmt.Errorf("invalid limit value: %w", err)
	}

	var offset int
	if len(c.Arguments) > 1 {
		offset, err = strconv.Atoi(c.Arguments[1])
		if err != nil {
			return fmt.Errorf("invalid offset value: %w", err)
		}
	}

	posts, err := s.DB.GetPostsForUser(ctx, gatordb.GetPostsForUserParams{
		UserID: c.User.ID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return fmt.Errorf("GetPostsForUser Failed: %w", err)
	}

	for _, post := range posts {
		fmt.Printf("* - %s\n", post.Title)
	}

	return nil
}

func GetUserMiddleware(handler func(context.Context, *state.State, Command) error) func(context.Context, *state.State, Command) error {
	return func(ctx context.Context, s *state.State, c Command) error {
		if s.Config.CurrentUserName == "" {
			return errors.New("You muts register or login in before run this command")
		}

		user, err := s.DB.GetUserByName(ctx, s.Config.CurrentUserName)
		if err != nil {
			return fmt.Errorf("Cannot get the user due to: %w", err)
		}

		c.User = user
		return handler(ctx, s, c)
	}
}

func checkUserExist(ctx context.Context, s *state.State, name string) (bool, error) {
	_, err := s.DB.GetUserByName(ctx, name)
	return !errors.Is(err, sql.ErrNoRows), err
}

func scrapeFeeds(ctx context.Context, s *state.State) error {

	nextFeed, err := s.DB.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("GetNextFeedToFetch Failed: %w", err)
	}

	feed, err := rss.FetchFeed(ctx, nextFeed.Url)
	if err != nil {
		return fmt.Errorf(" %s FetchFeed Fail: %w", nextFeed.Name, err)
	}

	err = s.DB.MarkFeedFetched(ctx, nextFeed.ID)
	if err != nil {
		return fmt.Errorf("MarkFeedFetched Failed: %w", err)
	}

	fmt.Printf("Feed %s fetched\n", feed.Channel.Title)
	for _, item := range feed.Channel.Item {
		pubTime, err := parseTime(item.PubDate)
		if err != nil {
			return fmt.Errorf("parseTiem Failed: %w", err)
		}

		hasDesc := len(item.Description) > 0

		post, err := s.DB.CreatePost(ctx,
			gatordb.CreatePostParams{
				ID:          uuid.New(),
				Title:       item.Title,
				FeedID:      nextFeed.ID,
				Url:         item.Link,
				PublishedAt: pubTime,
				Description: sql.NullString{
					Valid:  hasDesc,
					String: item.Description,
				},
			})

		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			fmt.Printf("%s already exists\n", item.Title)
			continue
		}

		if err != nil {
			return err
		}

		fmt.Printf("post created: %s: %s", post.ID, post.Title)
	}

	return nil
}

func parseTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC1123Z,
		time.Layout,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC3339,
		time.RFC3339Nano,
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format for %s, err: %w", s, err)
}
