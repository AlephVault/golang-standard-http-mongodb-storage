package dsl

// Auth is a table reference used for authentication
// purposes (API Keys).
type Auth struct {
	TableRef
}

// Prepare installs default values in the auth.
func (auth *Auth) Prepare() {
	if auth.Db == "" {
		auth.Db = "alephvault_http_storage"
	}
	if auth.Collection == "" {
		auth.Collection = "auth"
	}
}
