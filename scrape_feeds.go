package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Misaka8339-ctrl/gator/internal/database"

	"github.com/google/uuid"
)

func scrapeFeeds(ctx context.Context, s *state) error {
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("No feeds available. Add a feed with addfeed.")
			return nil
		}
		return fmt.Errorf("get next feed: %w", err)
	}

	if err := s.db.MarkFeedFetched(ctx, feed.ID); err != nil {
		return fmt.Errorf("mark feed fetched: %w", err)
	}

	fmt.Printf("Fetching feed %q: %s\n", feed.Name, feed.Url)

	rssFeed, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("fetch feed %q: %w", feed.Name, err)
	}

	var added, duplicates, failed int

	for _, item := range rssFeed.Channel.Item {
		if err := ctx.Err(); err != nil {
			return err
		}

		postURL := strings.TrimSpace(item.Link)
		if postURL == "" {
			log.Printf("Skipping post %q: missing URL", item.Title)
			failed++
			continue
		}

		publishedAt := parseFeedTime(item.PubDate)
		if !publishedAt.Valid && strings.TrimSpace(item.PubDate) != "" {
			log.Printf(
				"Post %q: unrecognized publication date %q; storing NULL",
				item.Title,
				item.PubDate,
			)
		}

		now := time.Now().UTC()

		rows, err := s.db.CreatePost(ctx, database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Title:     item.Title,
			Url:       postURL,
			Description: sql.NullString{
				String: item.Description,
				Valid:  strings.TrimSpace(item.Description) != "",
			},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			log.Printf("Save post %q: %v", postURL, err)
			failed++
			continue
		}

		if rows == 0 {
			duplicates++
		} else {
			added++
		}
	}

	fmt.Printf(
		"Finished feed %q: %d new, %d duplicates, %d failed/skipped\n",
		feed.Name,
		added,
		duplicates,
		failed,
	)

	return nil
}
