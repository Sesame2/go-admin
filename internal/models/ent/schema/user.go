package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StructTag(`json:"id,omitempty"`).
			Comment("用户 UUID"),

		field.String("username").
			NotEmpty().
			Unique(),

		field.String("email").
			NotEmpty().
			Unique(),

		field.String("password").
			NotEmpty(),

		field.String("role").
			NotEmpty().
			Default("user").
			Comment("用户角色"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
