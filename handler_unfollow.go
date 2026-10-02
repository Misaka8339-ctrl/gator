package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Misaka8339-ctrl/gator/internal/database"
)

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: unfollow <url>")
	}

	ctx := context.Background()

	feed, err := s.db.GetFeedByURL(ctx, cmd.args[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("feed not found")
		}
		return fmt.Errorf("get feed: %w", err)
	}

	err = s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("unfollow feed: %w", err)
	}

	fmt.Printf("User %s is no longer following %s\n", user.Name, feed.Name)
	return nil
}
