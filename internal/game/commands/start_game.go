package commands

type StartGameCommand struct {
	BaseCommand
}

func (c StartGameCommand) GetType() CommandType { return CommandStartGame }
