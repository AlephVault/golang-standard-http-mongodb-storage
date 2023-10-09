package dsl

// Global stands for settings for ALL the resources.
type Global struct {
	ListMaxResults uint `validate:"min:1"`
}
