package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/matttinkey/aggregotor/internal/commands"
	"github.com/matttinkey/aggregotor/internal/config"
	"github.com/matttinkey/aggregotor/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)
	s := commands.State{
		DB:  dbQueries,
		Cfg: &cfg,
	}

	cmds := commands.Commands{
		CmdMap: make(map[string]func(*commands.State, commands.Command) error),
	}

	commands.RegisterCommands(cmds)
	args := os.Args
	if len(args) < 2 {
		log.Fatal("no command given")
	}

	cmdName := args[1]
	var cmdArgs []string
	if len(args) > 2 {
		cmdArgs = args[2:]
	}

	cmd := commands.Command{
		Name: cmdName,
		Args: cmdArgs,
	}

	if err = cmds.Run(&s, cmd); err != nil {
		log.Fatal(err)
	}
}
