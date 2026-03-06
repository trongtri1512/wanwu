package orm

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
)

func (c *Client) CreateSkillConversation(ctx context.Context, conversation *model.SkillConversation) *errs.Status {
	if err := c.db.WithContext(ctx).Create(conversation).Error; err != nil {
		return toErrStatus("skill_conversation_create", err.Error())
	}
	return nil
}

func (c *Client) DeleteSkillConversation(ctx context.Context, conversationId, userId, orgId string) *errs.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithConversationId(conversationId),
		sqlopt.WithUserId(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Delete(&model.SkillConversation{}).Error; err != nil {
		return toErrStatus("skill_conversation_delete", err.Error())
	}
	return nil
}

func (c *Client) GetSkillConversationList(ctx context.Context, userId, orgId string, pageNo, pageSize int) ([]*model.SkillConversation, int64, *errs.Status) {
	var list []*model.SkillConversation
	var total int64

	if err := sqlopt.SQLOptions(
		sqlopt.WithUserId(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Model(&model.SkillConversation{}).
		Order("created_at desc").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, toErrStatus("skill_conversation_get_list", err.Error())
	}

	total = int64(len(list))

	return list, total, nil
}
