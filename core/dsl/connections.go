package dsl

import (
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"standard-http-mongodb-storage/core/connections"
	"strconv"
)

// ConnectionFields stands for the arguments needed
// to make a connection. This is just an alternative
// of arguments for the connection.
type ConnectionFields struct {
	Host     string
	Port     uint16
	Username string
	Password string
}

// Connection stands for the connection attempt, which
// might specify arguments (see: Args field) or might
// specify a direct URL (perhaps not mongodb://).
type Connection struct {
	Url     string
	Args    ConnectionFields
	Timeout int64
	client  *mongo.Client
}

// Prepare installs default values in the connection.
func (c *Connection) Prepare() {
	if c.Url == "" {
		if c.Args.Host == "" {
			if host, found := os.LookupEnv("MONGODB_HOST"); !found {
				c.Args.Host = "localhost"
			} else {
				c.Args.Host = host
			}
		}
		if c.Args.Port == 0 {
			if port, found := os.LookupEnv("MONGODB_PORT"); !found {
				c.Args.Port = 27017
			} else if portNumber, err := strconv.ParseUint(port, 10, 16); err != nil {
				c.Args.Port = 27017
			} else {
				c.Args.Port = uint16(portNumber)
			}
		}
		if c.Args.Username == "" {
			c.Args.Username = os.Getenv("MONGODB_USER")
		}
		if c.Args.Password == "" {
			c.Args.Password = os.Getenv("MONGODB_PASSWORD")
		}
	}
}

// Connect attempts a new connection.
func (c *Connection) Connect() (*mongo.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	options := connections.WithClientOptions().SetTimeout(c.Timeout)
	if c.Url != "" {
		return connections.ConnectWithURL(c.Url, options)
	} else {
		if cli, err := connections.ConnectWithFields(c.Args.Host, c.Args.Port, c.Args.Username, c.Args.Password, options); err != nil {
			c.client = nil
			return nil, err
		} else {
			return cli, nil
		}
	}
}

// Disconnect disconnects the current connection, if any.
func (c *Connection) Disconnect() error {
	if c.client == nil {
		return nil
	}

	options := connections.WithClientOptions().SetTimeout(c.Timeout)
	return connections.Disconnect(c.client, options)
}

// Client returns the underlying client, if any.
func (c *Connection) Client() *mongo.Client {
	return c.client
}
