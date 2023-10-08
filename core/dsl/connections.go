package dsl

// Connection stands to the operations that
// establish or close the connections.
type Connection struct {
	host     string
	port     string
	username string
	password string
	timeout  int64
}
