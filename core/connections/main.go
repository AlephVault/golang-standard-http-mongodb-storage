package connections

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// Makes a timeout context.
func makeTimeoutContext(timeout int64) (context.Context, func()) {
	return context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
}

// ConnectionOptions stands for the options related to the
// connection (either to connect, disconnect or interact).
type ConnectionOptions struct {
	timeout int64
}

// WithClientOptions creates a connection options.
func WithClientOptions() *ConnectionOptions {
	return &ConnectionOptions{timeout: 10}
}

// Connect makes a client out of the given input arguments.
func Connect(
	host string, port string, username string, password string, options *ConnectionOptions,
) (*mongo.Client, error) {
	if options == nil {
		options = WithClientOptions()
	}

	url := fmt.Sprintf("mongodb://%s:%s@%s:%s", host, port, username, password)
	timeoutContext, cancel := makeTimeoutContext(options.timeout)
	defer cancel()
	return mongo.Connect(timeoutContext, options.Client().ApplyURI(url))
}

// Disconnect disconnect.
func Disconnect(client *mongo.Client, options *ConnectionOptions) error {
	if options == nil {
		options = WithClientOptions()
	}

	timeoutContext, cancel := makeTimeoutContext(options.timeout)
	defer cancel()
	return client.Disconnect(timeoutContext)
}
