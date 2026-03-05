package model

type CustomSkill struct {
	ID         uint32 `gorm:"primarykey"`
	Name       string `gorm:"column:name;index:idx_custom_skill_name;comment:skill名称"`
	Avatar     string `gorm:"column:avatar;comment:skill头像"`
	Author     string `gorm:"column:author;comment:作者"`
	Desc       string `gorm:"column:desc;comment:skill描述"`
	ObjectPath string `gorm:"column:object_path;comment:skill数据minio对象路径(zip压缩包)"`
	Markdown   string `gorm:"column:markdown;type:text;comment:skill markdown内容"`
	SaveId     string `gorm:"column:save_id;index:idx_custom_skill_save_id;comment:保存id"`
	SourceType string `gorm:"column:source_type;comment:来源类型"`
	UserId     string `gorm:"column:user_id;index:idx_custom_skill_user_id;comment:用户id"`
	OrgId      string `gorm:"column:org_id;index:idx_custom_skill_org_id;comment:组织id"`
	CreatedAt  int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt  int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
