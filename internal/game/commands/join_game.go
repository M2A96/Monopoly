package commands

type JoinGameCommand struct {
	BaseCommand
	Name string
}

func (c JoinGameCommand) GetType() CommandType { return CommandJoinGame }
