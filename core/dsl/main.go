package dsl

// Settings stands for the main entry point of our DSL.
type Settings struct {
	Debug      bool
	Connection Connection
	Global     Global `validate:"dive"`
	Auth       Auth   `validate:"dive"`
}

// Prepare prepares the default values of all the members.
func (settings *Settings) Prepare() {
	settings.Connection.Prepare()
	settings.Auth.Prepare()
}
