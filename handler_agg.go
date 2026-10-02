package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: agg <time_between_reqs>")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", cmd.args[0], err)
	}

	if timeBetweenRequests <= 0 {
		return fmt.Errorf("time_between_reqs must be greater than zero")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	for {
		if ctx.Err() != nil {
			fmt.Println("\nAggregation stopped")
			return nil
		}

		if err := scrapeFeeds(ctx, s); err != nil {
			if ctx.Err() != nil {
				fmt.Println("\nAggregation stopped")
				return nil
			}
			fmt.Fprintln(os.Stderr, "scrape error:", err)
		}

		select {
		case <-ctx.Done():
			fmt.Println("\nAggregation stopped")
			return nil
		case <-ticker.C:
			// 到达下一个定时点，进入下一轮。
		}
	}
}
