package tracking

// PublishForTest exposes notifySubscribers to tests in other packages
// so they can drive the fan-out path without running SQL. Not safe
// for production callers — Record() is the only sanctioned publisher.
func PublishForTest(rec *CommandRecord) {
	notifySubscribers(rec)
}
