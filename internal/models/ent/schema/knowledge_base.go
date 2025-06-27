package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type KnowledgeBase struct {
	ent.Schema
}

func (KnowledgeBase) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StructTag(`json:"id":,omitempty`).
			Comment("数据库ID"),
		field.String("name").
			NotEmpty().
			Comment("知识库名称"),
		field.String("dataset_id").
			Optional().
			Comment("关丽娜的数据库ID"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更改时间"),
	}

}

// Edges of the KnowledgeBase.
func (KnowledgeBase) Edges() []ent.Edge {
	return nil
	// return []ent.Edge{
	// 	// 可以在此添加与其他实体的关系
	// 	// 例如与 Dataset 的关系
	// 	// edge.To("dataset", Dataset.Type).
	// 	//    Unique().
	// 	//    Field("dataset_id").
	// 	//    Comment("关联的数据集"),
	// }
}
