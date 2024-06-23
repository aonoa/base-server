package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Rule holds the schema definition for the Rule entity.
type Rule struct {
	ent.Schema
}

// Fields of the Rule.
func (Rule) Fields() []ent.Field {
	return []ent.Field{
		field.String("ptype"),
		field.String("v0"),
		field.String("v1"),
		field.String("v2"),
		field.String("v3"),
		field.String("v4"),
		field.String("v5"),
	}
}

// Mixin .
func (Rule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		// Or, mixin.CreateTime only for create_time
		// and mixin.UpdateTime only for update_time.
	}
}

// Edges of the Rule.
func (Rule) Edges() []ent.Edge {
	return nil
}
