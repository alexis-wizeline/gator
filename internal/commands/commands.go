package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/alexis-wizeline/gator/internal/gatordb"
	"github.com/alexis-wizeline/gator/internal/rss"
	"github.com/alexis-wizeline/gator/internal/state"
)

type Commands struct {
	handlers map[string]func(context.Context, *state.State, Command) error
}

func (c *Commands) Run(ctx context.Context, s *state.State, cmd Command) error {
	handler, ok := c.handlers[cmd.Name]
	if !ok {
		return errors.New("unknow command")
	}
	return handler(ctx, s, cmd)
}

func (c *Commands) Register(name string, f func(context.Context, *state.State, Command) error) {
	c.handlers[name] = f
}

func GatorCommands() *Commands {
	handlers := make(map[string]func(context.Context, *state.State, Command) error)
	return &Commands{
		handlers,
	}
}

type Command struct {
	Name      string
	Arguments []string
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
		return err
	}

	err = s.Config.SetUser(username)
	if err != nil {
		return err
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
		return err
	}

	currentTime := time.Now()
	u, err := s.DB.CreateUser(ctx, gatordb.CreateUserParams{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return err
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
		return err
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

func HandleAgg(ctx context.Context, _ *state.State, c Command) error {
	url := "https://www.wagslane.dev/index.xml"
	if len(c.Arguments) > 0 {
		url = c.Arguments[0]
	}

	feed, err := rss.FetchFeed(ctx, url)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", feed)

	return nil
}

func checkUserExist(ctx context.Context, s *state.State, name string) (bool, error) {
	_, err := s.DB.GetUserByName(ctx, name)
	return !errors.Is(err, sql.ErrNoRows), err
}
