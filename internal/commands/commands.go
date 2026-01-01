package commands

import (
	"fmt"

	"github.com/matttinkey/aggregotor/internal/config"
	"github.com/matttinkey/aggregotor/internal/database"
)

type State struct {
	Cfg *config.Config
	DB  *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	CmdMap map[string]func(*State, Command) error
}

func (c *Commands) Run(s *State, cmd Command) error {
	cmdFunc, ok := c.CmdMap[cmd.Name]
	if !ok {
		return fmt.Errorf("command \"%s\" not found", cmd.Name)
	}

	return cmdFunc(s, cmd)
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.CmdMap[name] = f
}

func RegisterCommands(cmds Commands) {
	cmds.Register("login", handlerLogin)
	cmds.Register("register", handlerRegister)
	cmds.Register("reset", handlerReset)
	cmds.Register("users", handlerGetUsers)
	cmds.Register("agg", handlerAgg)
	cmds.Register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.Register("feeds", hanlderFeeds)
	cmds.Register("follow", middlewareLoggedIn(handlerFollow))
	cmds.Register("following", handlerFollowing)
	cmds.Register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.Register("browse", middlewareLoggedIn(handlerBrowse))
}
