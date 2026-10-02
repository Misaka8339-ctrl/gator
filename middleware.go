package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Misaka8339-ctrl/gator/internal/database"
)

func middlewareLoggedIn(
	handler func(s *state, cmd command, user database.User) error,
) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(
			context.Background(),
			s.cfg.CurrentUserName,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("current user does not exist; register or login first")
			}
			return fmt.Errorf("get current user: %w", err)
		}

		return handler(s, cmd, user)
	}
}
