package dsl

const DefaultListMaxSize int64 = 20

// Global stands for settings for ALL the resources.
type Global struct {
	ListMaxResults int64
}

// Prepare installs the default values in the global settings.
func (global *Global) Prepare() {
	if global.ListMaxResults == 0 {
		global.ListMaxResults = 1
	}
}
