package main

import (
	"fmt"

	"github.com/Misaka8339-ctrl/gator/internal/config"
	"github.com/Misaka8339-ctrl/gator/internal/database"
)

// 所有命令共享的应用状态。
type state struct {
	db  *database.Queries
	cfg *config.Config
}

// 用户输入的一条命令。
type command struct {
	name string
	args []string
}

// 命令名称 → 对应的处理函数。
type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(
	name string,
	f func(*state, command) error,
) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}

	return handler(s, cmd)
}
