package dsl

// TableRef stands for the db and name of a given
// table (e.g. "global" and "auth").
type TableRef struct {
	Db         string `validate:"mdb-name"`
	Collection string `validate:"mdb-name"`
}
