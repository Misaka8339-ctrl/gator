package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Misaka8339-ctrl/gator/internal/database"
)

func handlerBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 1 {
		return fmt.Errorf("usage: browse [limit]")
	}

	limit := int32(2)

	if len(cmd.args) == 1 {
		n, err := strconv.ParseInt(cmd.args[0], 10, 32)
		if err != nil || n <= 0 {
			return fmt.Errorf("limit must be a positive 32-bit integer")
		}
		limit = int32(n)
	}

	posts, err := s.db.GetPostsForUser(
		context.Background(),
		database.GetPostsForUserParams{
			UserID: user.ID,
			Limit:  limit,
		},
	)
	if err != nil {
		return fmt.Errorf("get posts for user: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts found. Follow a feed and let agg fetch its posts.")
		return nil
	}

	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("Feed: %s\n", post.FeedName)

		if post.PublishedAt.Valid {
			fmt.Printf(
				"Published: %s\n",
				post.PublishedAt.Time.UTC().Format(time.RFC3339),
			)
		} else {
			fmt.Println("Published: unknown")
		}

		fmt.Printf("URL: %s\n", post.Url)

		if post.Description.Valid {
			fmt.Printf("Description: %s\n", post.Description.String)
		}

		fmt.Println()
	}
	return nil
}
