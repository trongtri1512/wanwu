package assistant

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) CreateSkillConversation(ctx context.Context, req *assistant_service.CreateSkillConversationReq) (*assistant_service.CreateSkillConversationResp, error) {
	conversationId := uuid.New().String()

	err := s.cli.CreateSkillConversation(ctx, &model.SkillConversation{
		ConversationId: conversationId,
		Title:          req.Title,
		UserId:         req.Identity.UserId,
		OrgId:          req.Identity.OrgId,
	})
	if err != nil {
		return nil, errStatus(errs.Code_SkillConversationCreateErr, err)
	}
	return &assistant_service.CreateSkillConversationResp{
		ConversationId: conversationId,
	}, nil
}

func (s *Service) DeleteSkillConversation(ctx context.Context, req *assistant_service.DeleteSkillConversationReq) (*emptypb.Empty, error) {
	err := s.cli.DeleteSkillConversation(ctx, req.ConversationId, req.Identity.UserId, req.Identity.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_SkillConversationDeleteErr, err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) GetSkillConversationList(ctx context.Context, req *assistant_service.GetSkillConversationListReq) (*assistant_service.GetSkillConversationListResp, error) {
	list, total, err := s.cli.GetSkillConversationList(ctx, req.Identity.UserId, req.Identity.OrgId, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetCustomMCPListErr, err)
	}

	respList := make([]*assistant_service.SkillConversationItem, 0, len(list))
	for _, item := range list {
		respList = append(respList, &assistant_service.SkillConversationItem{
			ConversationId: item.ConversationId,
			Title:          item.Title,
			CreatedAt:      item.CreatedAt,
		})
	}

	return &assistant_service.GetSkillConversationListResp{
		List:  respList,
		Total: total,
	}, nil
}
