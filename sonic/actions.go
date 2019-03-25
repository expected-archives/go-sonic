package sonic

type Action string

const (
	Consolidate Action = "consolidate"
)

func IsActionValid(action Action) bool {
	return action == Consolidate
}
