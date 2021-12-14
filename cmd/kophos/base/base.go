package base

type Command struct {
	Run func(args []string)
}
