package commands

type CreateGameCommand struct {
	BaseCommand
	Name string
}

func (c CreateGameCommand) GetType() CommandType { return CommandCreateGame }
