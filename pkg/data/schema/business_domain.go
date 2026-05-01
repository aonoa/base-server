package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// BusinessDomain holds the schema definition for platform business domains.
type BusinessDomain struct {
	ent.Schema
}

func (BusinessDomain) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("平台业务域注册表"),
		entsql.Annotation{
			Table: "sys_business_domain",
		},
	}
}

func (BusinessDomain) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Comment("数据唯一标识"),
		field.String("code").
			Unique().
			Comment("业务域编码"),
		field.String("name").Comment("业务域名称"),
		field.String("owner_service").Comment("主服务编码"),
		field.String("org_model_type").Comment("组织模型类型"),
		field.String("auth_scope_type").Comment("权限范围类型"),
		field.Bool("status").Default(true).Comment("0-禁用，1-启用"),
		field.String("description").Default("").Comment("描述"),
		field.String("meta_json").Default("").Comment("扩展元数据"),
	}
}

func (BusinessDomain) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (BusinessDomain) Edges() []ent.Edge {
	return nil
}
