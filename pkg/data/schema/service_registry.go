package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// ServiceRegistry holds the schema definition for platform service registration.
type ServiceRegistry struct {
	ent.Schema
}

func (ServiceRegistry) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("平台服务注册表"),
		entsql.Annotation{
			Table: "sys_service_registry",
		},
	}
}

func (ServiceRegistry) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Comment("数据唯一标识"),
		field.String("service_code").
			Unique().
			Comment("服务编码"),
		field.String("service_name").Comment("服务名称"),
		field.String("domain_code").Comment("所属业务域"),
		field.String("http_prefix").Default("").Comment("HTTP 前缀"),
		field.String("grpc_service").Default("").Comment("gRPC 服务名"),
		field.Bool("status").Default(true).Comment("0-禁用，1-启用"),
		field.Bool("projection_enabled").Default(false).Comment("是否启用权限投影"),
		field.String("description").Default("").Comment("描述"),
	}
}

func (ServiceRegistry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (ServiceRegistry) Edges() []ent.Edge {
	return nil
}
