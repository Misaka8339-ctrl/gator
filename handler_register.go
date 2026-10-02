package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Misaka8339-ctrl/gator/internal/database"

	"github.com/google/uuid"
)

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: register <username>")
	}

	name := cmd.args[0]
	now := time.Now().UTC()

	user, err := s.db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Name:      name,
		},
	)
	if err != nil {
		return fmt.Errorf("create user %q: %w", name, err)
	}

	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf(
			"user was created, but updating the current user failed: %w",
			err,
		)
	}

	fmt.Printf("User %s created successfully\n", user.Name)
	fmt.Printf("%+v\n", user)

	return nil
}
