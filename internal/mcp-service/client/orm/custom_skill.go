package orm

import (
	"context"
	"errors"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

func (c *Client) CreateCustomSkill(ctx context.Context, customSkill *model.CustomSkill) (string, *err_code.Status) {
	status := c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		if err := tx.WithContext(ctx).Create(customSkill).Error; err != nil {
			return toErrStatus("mcp_custom_skill_create", err.Error())
		}
		return nil
	})

	if status != nil {
		return "", status
	}

	return util.Int2Str(customSkill.ID), nil
}

func (c *Client) DeleteCustomSkill(ctx context.Context, skillId string) *err_code.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(util.MustU32(skillId)),
	).Apply(c.db).WithContext(ctx).Delete(&model.CustomSkill{}).Error; err != nil {
		return toErrStatus("mcp_custom_skill_delete", err.Error())
	}
	return nil
}

func (c *Client) GetCustomSkill(ctx context.Context, skillId string) (*model.CustomSkill, *err_code.Status) {
	var cs model.CustomSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(util.MustU32(skillId)),
	).Apply(c.db).WithContext(ctx).First(&cs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toErrStatus("mcp_custom_skill_not_found", skillId)
		}
		return nil, toErrStatus("mcp_custom_skill_get", skillId, err.Error())
	}
	return &cs, nil
}

func (c *Client) GetCustomSkillList(ctx context.Context, userId, orgId, name string) ([]*model.CustomSkill, int64, *err_code.Status) {
	var list []*model.CustomSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
		sqlopt.LikeName(name),
	).Apply(c.db).WithContext(ctx).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, toErrStatus("mcp_custom_skill_list", err.Error())
	}

	return list, int64(len(list)), nil
}

func (c *Client) GetCustomSkillBySaveIds(ctx context.Context, saveIds []string) ([]*model.CustomSkill, *err_code.Status) {
	var list []*model.CustomSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithCustomSkillSaveIds(saveIds),
	).Apply(c.db).WithContext(ctx).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, toErrStatus("mcp_custom_skill_get_by_save_ids", err.Error())
	}

	return list, nil
}
