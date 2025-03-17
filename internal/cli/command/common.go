package command

// Command Common interface for all available cli commands.
type Command interface {
	// Run Main command, args — all arguments from cmd except just first.
	Run(args []string) error
}
