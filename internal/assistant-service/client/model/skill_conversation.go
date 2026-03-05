package model

type SkillConversation struct {
	ID             uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:'id'"`
	ConversationId string `gorm:"uniqueIndex:idx_unique_conversation_id;column:conversation_id;type:varchar(255);not null;comment:'conversation id'"`
	Title          string `gorm:"column:title;type:varchar(255);comment:'conversation title'"`
	UserId         string `gorm:"column:user_id;index:idx_user_id_org_id,priority:1;type:varchar(64);not null;comment:'user id'"`
	OrgId          string `gorm:"column:org_id;index:idx_user_id_org_id,priority:2;type:varchar(64);not null;comment:'org id'"`
	CreatedAt      int64  `gorm:"autoCreateTime:milli;comment:create time"`
	UpdatedAt      int64  `gorm:"autoUpdateTime:milli;comment:update time"`
}
