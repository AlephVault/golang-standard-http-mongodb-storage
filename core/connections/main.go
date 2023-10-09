package connections

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	mongodboptions "go.mongodb.org/mongo-driver/mongo/options"
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

// SetTimeout sets the timeout for the connection options.
func (options *ConnectionOptions) SetTimeout(timeout int64) *ConnectionOptions {
	if timeout > 0 {
		options.timeout = timeout
	}
	return options
}

// ConnectWithFields makes a client out of the given input arguments.
func ConnectWithFields(
	host string, port uint16, username string, password string, options *ConnectionOptions,
) (*mongo.Client, error) {
	return ConnectWithURL(fmt.Sprintf("mongodb://%s:%s@%s:%s", host, port, username, password), options)
}

// ConnectWithURL makes a client out of the given input URL.
func ConnectWithURL(url string, options *ConnectionOptions) (*mongo.Client, error) {
	if options == nil {
		options = WithClientOptions()
	}

	timeoutContext, cancel := makeTimeoutContext(options.timeout)
	defer cancel()
	return mongo.Connect(timeoutContext, mongodboptions.Client().ApplyURI(url))
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
