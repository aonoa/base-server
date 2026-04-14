package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Annotations of the User.
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("用户信息表"),
		entsql.Annotation{
			Table: "sys_user",
		},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("username").Comment("用户名"),
		field.String("password").Comment("密码"),
		field.String("nickname").Optional().Comment("昵称"),
		field.String("email").Optional().Comment("邮箱"),
		field.Int8("status").Comment("0-锁定，1-正常"),
		field.String("avatar").Comment("头像"),
		field.String("desc").Comment("备注"),
		field.String("extension").Comment("扩展信息"),
	}
}

// Mixin .
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
