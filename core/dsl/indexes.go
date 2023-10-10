package dsl

// Index is the description of a MongoDB index.
type Index struct {
	Unique bool
	Fields []string `validate:"required,dive,required,mdb-index-entry"`
}
