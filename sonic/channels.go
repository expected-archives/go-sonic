package sonic

// Channel refer to the list of channels available.
type Channel string

const (
	// Search  is used for querying the search index.
	Search Channel = "search"

	// Ingest is used for altering the search index (push, pop and flush).
	Ingest Channel = "ingest"

	// Control is used for administration purposes.
	Control Channel = "control"
)

// IsChannelValid check if the parameter is a valid channel.
func IsChannelValid(ch Channel) bool {
	return ch == Search || ch == Ingest || ch == Control
}
