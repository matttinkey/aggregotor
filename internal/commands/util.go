package commands

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/matttinkey/aggregotor/internal/database"
	"github.com/matttinkey/aggregotor/internal/rss"
)

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		ctx := context.Background()
		user, err := s.DB.GetUser(ctx, s.Cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func scrapeFeeds(s *State) error {
	ctx := context.Background()
	feedInfo, err := s.DB.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}

	if err := s.DB.MarkFeedFetched(ctx, feedInfo.ID); err != nil {
		return err
	}

	feed, err := rss.FetchFeed(ctx, feedInfo.Url)
	if err != nil {
		return err
	}

	for _, post := range feed.Channel.Item {
		// fmt.Println(post.Title)
		pubTime, err := time.Parse(time.RFC1123Z, post.PubDate)
		if err != nil {
			return err
		}

		params := database.CreatePostParams{
			ID:        int32(uuid.New().ID()),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Title:     post.Title,
			Url:       post.Link,
			Description: sql.NullString{
				String: post.Description,
			},
			PublishedAt: pubTime,
			FeedID: sql.NullInt32{
				Int32: feedInfo.ID,
				Valid: true,
			},
		}
		_, err = s.DB.CreatePost(ctx, params)
		if err == nil {
			fmt.Printf("saved post \"%v\" to database\n", post.Title)
		}
	}

	return nil
}
