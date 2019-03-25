package sonic

type Channel string

const (
	Search  Channel = "search"
	Ingest  Channel = "ingest"
	Control Channel = "control"
)

func IsChannelValid(ch Channel) bool {
	return ch == Search || ch == Ingest || ch == Control
}
