package proposal

const (
	ClosedState  State = "closed"
	FinalState   State = "final"
	ActiveState  State = "active"
	PendingState State = "pending"
)

type State string
