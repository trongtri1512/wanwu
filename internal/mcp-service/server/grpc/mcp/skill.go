package mcp

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) CustomSkillCreate(ctx context.Context, req *mcp_service.CustomSkillCreateReq) (*mcp_service.CustomSkillCreateResp, error) {
	skillId, err := s.cli.CreateCustomSkill(ctx, &model.CustomSkill{
		Name:       req.Name,
		Avatar:     req.Avatar,
		Author:     req.Author,
		Desc:       req.Desc,
		ObjectPath: req.ObjectPath,
		Markdown:   req.Markdown,
		SaveId:     req.SaveId,
		SourceType: req.SourceType,
		UserId:     req.Identity.UserId,
		OrgId:      req.Identity.OrgId,
	})
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}

	return &mcp_service.CustomSkillCreateResp{SkillId: skillId}, nil
}

func (s *Service) CustomSkillDelete(ctx context.Context, req *mcp_service.CustomSkillDeleteReq) (*emptypb.Empty, error) {
	err := s.cli.DeleteCustomSkill(ctx, req.SkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) CustomSkillGet(ctx context.Context, req *mcp_service.CustomSkillGetReq) (*mcp_service.CustomSkill, error) {
	customSkill, err := s.cli.GetCustomSkill(ctx, req.SkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}

	return toCustomSkillInfo(customSkill), nil
}

func (s *Service) CustomSkillGetList(ctx context.Context, req *mcp_service.CustomSkillGetListReq) (*mcp_service.CustomSkillGetListResp, error) {
	customSkills, total, err := s.cli.GetCustomSkillList(ctx, req.Identity.UserId, req.Identity.OrgId, req.Name)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}

	customSkillList := make([]*mcp_service.CustomSkill, 0, len(customSkills))
	for _, customSkill := range customSkills {
		customSkillList = append(customSkillList, toCustomSkillInfo(customSkill))
	}

	return &mcp_service.CustomSkillGetListResp{
		List:  customSkillList,
		Total: total,
	}, nil
}

func (s *Service) CustomSkillGetBySaveIds(ctx context.Context, req *mcp_service.CustomSkillGetBySaveIdsReq) (*mcp_service.CustomSkillSaveIdsResp, error) {
	customSkills, err := s.cli.GetCustomSkillBySaveIds(ctx, req.SaveIds)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}

	saveIds := make([]string, 0, len(customSkills))
	for _, customSkill := range customSkills {
		saveIds = append(saveIds, customSkill.SaveId)
	}

	return &mcp_service.CustomSkillSaveIdsResp{
		SaveIds: saveIds,
	}, nil
}

func toCustomSkillInfo(customSkill *model.CustomSkill) *mcp_service.CustomSkill {
	if customSkill == nil {
		return nil
	}
	return &mcp_service.CustomSkill{
		SkillId:    util.Int2Str(customSkill.ID),
		Name:       customSkill.Name,
		Avatar:     customSkill.Avatar,
		Author:     customSkill.Author,
		Desc:       customSkill.Desc,
		ObjectPath: customSkill.ObjectPath,
		Markdown:   customSkill.Markdown,
		CreatedAt:  customSkill.CreatedAt,
		UpdatedAt:  customSkill.UpdatedAt,
	}
}
