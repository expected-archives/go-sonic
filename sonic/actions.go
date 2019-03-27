package sonic

// Action refer to list of actions for TRIGGER command.
type Action string

const (
	// Consolidate action is not detailed in the sonic protocol.
	Consolidate Action = "consolidate"
)

// IsActionValid check if the action passed in parameter is valid.
// Mean that TRIGGER command can handle it.
func IsActionValid(action Action) bool {
	return action == Consolidate
}
