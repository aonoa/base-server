package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

type SysLogRecord struct {
	ent.Schema
}

func (SysLogRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("系统日志表"),
		entsql.Annotation{
			Table: "sys_log",
		},
	}
}

func (SysLogRecord) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Immutable().
			Comment("数据唯一标识"),
		field.String("user_id").Comment("用户id"),
		field.String("user_name").Comment("用户名"),
		field.Bool("is_login").Comment("是否登录日志"),
		field.String("session_id").Comment("会话ID"),
		field.String("method").Comment("请求方式GET|POST|PUT|DELETE|OPTIONS..."),
		field.String("path").Comment("请求地址"),
		field.Time("request_time").Comment("请求时间"),
		field.String("ip_address").Comment("IP地址"),
		field.String("ip_location").Comment("IP归属地"),
		field.Int64("latency").Comment("响应耗时ms"),
		field.String("os").Comment("平台"),
		field.String("browser").Comment("浏览器"),
		field.String("user_agent").Comment("访问代理"),
		field.String("header").Comment("请求头"),
		field.String("get_params").Comment("Get参数"),
		field.String("post_data").Comment("Post数据"),
		field.Int32("res_code").Comment("响应状态码"),
		field.String("reason").Comment("错误原因"),
		field.Bool("res_status").Comment("请求状态"),
		field.String("stack").Comment("错误堆栈"),
		field.Time("create_time").
			Default(time.Now).
			Immutable(),
	}
}

func (SysLogRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id"),
		index.Fields("user_id"),
		index.Fields("path"),
		index.Fields("request_time"),
		index.Fields("ip_address"),
		index.Fields("latency"),
		// 只索引 is_login 为 true 的行
		index.Fields("is_login").
			Annotations(
				entsql.IndexWhere("is_login = true"),
			),
		index.Fields("is_login", "user_name").
			Annotations(
				entsql.IndexWhere("is_login = true"),
			),
		index.Fields("res_code").
			Annotations(
				entsql.IndexWhere("res_code != 200"),
			),
		index.Fields("res_status").
			Annotations(
				entsql.IndexWhere("res_status = false"),
			),
	}
}
