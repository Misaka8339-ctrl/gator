package main

import (
	"context"
	"fmt"
)

func handlerReset(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: reset")
	}

	if err := s.db.DeleteUsers(context.Background()); err != nil {
		return fmt.Errorf("reset database: %w", err)
	}

	fmt.Println("Database reset successfully")
	return nil
}
