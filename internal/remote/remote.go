package remote

const fmtErrPreFlight = "Prerequisite checks and actions failed for '%s' with error: %s"

type Remote interface {
	Init() error
	Sync() error
	Publish(snapshot string) error
}
