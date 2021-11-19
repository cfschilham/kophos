package command

type Command struct {
	Run func(args []string)
}
