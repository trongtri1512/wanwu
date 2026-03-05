package orm

import (
	"context"
	"errors"
	"regexp"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

type Client struct {
	db *gorm.DB
}

func NewClient(db *gorm.DB) (*Client, error) {
	// auto migrate
	if err := db.AutoMigrate(
		model.Assistant{},
		model.Conversation{},
		model.AssistantWorkflow{},
		model.AssistantMCP{},
		model.AssistantTool{},
		model.CustomPrompt{},
		model.AssistantSnapshot{},
		model.MultiAgentRelation{},
		model.SkillConversation{},
	); err != nil {
		return nil, err
	}

	if err := initAssistantUUID(db); err != nil {
		return nil, err
	}

	if err := initConversationType(db); err != nil {
		return nil, err
	}

	return &Client{
		db: db,
	}, nil
}

func initAssistantUUID(dbClient *gorm.DB) error {
	const batchSize = 100

	for {
		var ids []uint32
		if err := dbClient.Model(&model.Assistant{}).Select("id").Where("uuid = ? OR uuid IS NULL", "").Limit(batchSize).Find(&ids).Error; err != nil {
			return err
		}

		if len(ids) == 0 {
			break
		}

		caseWhen := "CASE id "
		var args []interface{}
		for _, id := range ids {
			caseWhen += "WHEN ? THEN ? "
			args = append(args, id, util.NewID())
		}
		caseWhen += "END"

		if err := dbClient.Model(&model.Assistant{}).
			Where("id IN ?", ids).
			UpdateColumn("uuid", gorm.Expr(caseWhen, args...)).Error; err != nil {
			log.Errorf("init assistant uuid batch update error: %v", err)
			return err
		}
	}

	return nil
}

func initConversationType(dbClient *gorm.DB) error {
	const batchSize = 100
	numericRegex := regexp.MustCompile(`^\d+$`)

	for {
		var conversations []model.Conversation
		if err := dbClient.Model(&model.Conversation{}).
			Select("id", "user_id").
			Where("conversation_type = ? OR conversation_type IS NULL", "").
			Limit(batchSize).
			Find(&conversations).Error; err != nil {
			return err
		}

		if len(conversations) == 0 {
			break
		}

		caseWhen := "CASE id "
		var args []interface{}
		var ids []uint32

		for _, conv := range conversations {
			ids = append(ids, conv.ID)
			newType := constant.ConversationTypeWebURL
			if numericRegex.MatchString(conv.UserId) {
				newType = constant.ConversationTypePublished
			}
			caseWhen += "WHEN ? THEN ? "
			args = append(args, conv.ID, newType)
		}
		caseWhen += "END"

		if err := dbClient.Model(&model.Conversation{}).
			Where("id IN ?", ids).
			UpdateColumn("conversation_type", gorm.Expr(caseWhen, args...)).Error; err != nil {
			log.Errorf("init conversation type batch update error: %v", err)
			return err
		}
	}

	return nil
}

func (c *Client) transaction(ctx context.Context, fc func(tx *gorm.DB) *err_code.Status) *err_code.Status {
	var status *err_code.Status
	_ = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if status = fc(tx); status != nil {
			return errors.New(status.String())
		}
		return nil
	})
	return status
}

func toErrStatus(code string, args ...string) *err_code.Status {
	return &err_code.Status{
		TextKey: code,
		Args:    args,
	}
}

func ErrCode(code err_code.Code) error {
	return grpc_util.ErrorStatusWithKey(code, "")
}
