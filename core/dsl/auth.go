package dsl

type Auth struct {
	Db         string `validate:"mdb-name"`
	Collection string `validate:"mdb-name"`
}
