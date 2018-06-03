package product

import "github.com/a-trium/domain-driven-design/implementation-duk/service-gateway/internal/domain"

type Tag struct {
	domain.BaseModel
	Name             string   `gorm:"column:name; type:varchar(20); not null; unique; index;"`
}

func (Tag) TableName() string {
	return "tag"
}