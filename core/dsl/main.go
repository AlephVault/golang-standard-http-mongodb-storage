package dsl

// Settings stands for the main entry point of our DSL.
type Settings struct {
	Debug      bool
	Connection Connection
	Global     Global              `validate:"dive"`
	Auth       Auth                `validate:"dive"`
	Resources  map[string]Resource `validate:"dive,keys,mdb-name,endkeys,dive"`
}

// Prepare prepares the default values of all the members.
func (settings *Settings) Prepare() *Settings {
	settings.Global.Prepare()
	settings.Connection.Prepare()
	settings.Auth.Prepare()
	return settings
}
