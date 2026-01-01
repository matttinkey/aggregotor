package commands

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/matttinkey/aggregotor/internal/database"
)

func handlerLogin(s *State, cmd Command) error {
	if cmd.Args == nil {
		return fmt.Errorf("no arguments given")
	}

	name := cmd.Args[0]
	ctx := context.Background()
	_, err := s.DB.GetUser(ctx, name)
	if err != nil {
		return fmt.Errorf("no user found with username '%v'", name)
	}

	if err := s.Cfg.SetUser(name); err != nil {
		return err
	}

	fmt.Println("username has been set")
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if cmd.Args == nil {
		return fmt.Errorf("no arguments given")
	}

	name := cmd.Args[0]

	ctx := context.Background()
	_, err := s.DB.GetUser(ctx, name)
	if err == nil {
		return fmt.Errorf("username %v already exists", name)
	}

	params := database.CreateUserParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
	}

	_, err = s.DB.CreateUser(ctx, params)
	if err != nil {
		return err
	}

	if err := s.Cfg.SetUser(name); err != nil {
		return err
	}

	fmt.Printf("user '%v' created", name)
	return nil
}

func handlerReset(s *State, cmd Command) error {
	if err := s.DB.Reset(context.Background()); err != nil {
		fmt.Println("DB reset was not successful")
		return err
	}

	fmt.Println("DB reset was successful")
	return nil
}

func handlerGetUsers(s *State, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not retrieve users: %v", err)
	}

	if len(users) == 0 {
		fmt.Println("No users found")
		return nil
	}

	for _, user := range users {
		name := user.Name
		if name == s.Cfg.CurrentUserName {
			name += " (current)"
		}
		fmt.Printf("* %v\n", name)
	}
	return nil
}

func handlerAgg(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("one argument required for agg")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
}

func handlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("two arguments required for addfeed")
	}

	name := cmd.Args[0]
	url := cmd.Args[1]
	ctx := context.Background()
	feed := database.CreateFeedParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	_, err := s.DB.CreateFeed(ctx, feed)
	if err != nil {
		return fmt.Errorf("could not add feed: %v", err)
	}

	feed_follow := database.CreateFeedFollowParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    feed.UserID,
		FeedID:    feed.ID,
	}
	_, err = s.DB.CreateFeedFollow(ctx, feed_follow)
	if err != nil {
		return err
	}

	fmt.Println(feed)
	return nil
}

func hanlderFeeds(s *State, cmd Command) error {
	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("%v:\n  url: %v\n  user: %v\n\n", feed.Name, feed.Url, feed.Name_2)
	}
	return nil
}

func handlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("follow command only takes one argument")
	}

	url := cmd.Args[0]
	ctx := context.Background()
	ids, err := s.DB.GetIDsFromUrl(ctx, url)
	if err != nil {
		return err
	}

	params := database.CreateFeedFollowParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    ids.FeedID,
	}

	result, err := s.DB.CreateFeedFollow(ctx, params)
	if err != nil {
		return err
	}

	fmt.Printf("feed: %v\nuser: %v\n", result.FeedName, result.UserName)
	return nil
}

func handlerFollowing(s *State, cmd Command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("following command takes no arguments")
	}

	name := s.Cfg.CurrentUserName
	ctx := context.Background()
	result, err := s.DB.GetFeedFollowsForUser(ctx, name)
	if err != nil {
		return err
	}

	fmt.Printf("%v is following these feeds:\n", name)
	for _, info := range result {
		fmt.Printf("- %v\n", info.FeedName)
	}
	return nil
}

func handlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("unfollow command only takes one argument")
	}

	ctx := context.Background()
	url := cmd.Args[0]
	feed, err := s.DB.GetFeedFromURL(ctx, url)
	if err != nil {
		return err
	}

	params := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	if err := s.DB.DeleteFeedFollow(ctx, params); err != nil {
		return err
	}

	fmt.Printf("user %v  has successfully unfollowed %v", user.Name, url)
	return nil
}

func handlerBrowse(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("browse command can only take one optional argument")
	}

	limit := 2
	if len(cmd.Args) == 1 {
		arg, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return err
		}
		limit = arg
	}

	ctx := context.Background()
	params := database.GetPostsForUserParams{
		Name:  user.Name,
		Limit: int32(limit),
	}
	posts, err := s.DB.GetPostsForUser(ctx, params)
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Println(post.Title)
	}

	return err
}
