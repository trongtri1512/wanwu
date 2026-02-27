package service

import (
	"strings"

	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func GetToolSquareDetail(ctx *gin.Context, userID, orgID, toolSquareID string) (*response.ToolSquareDetail, error) {
	resp, err := mcp.GetSquareTool(ctx.Request.Context(), &mcp_service.GetSquareToolReq{
		ToolSquareId: toolSquareID,
		Identity: &mcp_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	if err != nil {
		return nil, err
	}
	return toToolSquareDetail(ctx, resp), nil
}

func GetToolSquareList(ctx *gin.Context, userID, orgID, name string) (*response.ListResult, error) {
	resp, err := mcp.GetSquareToolList(ctx.Request.Context(), &mcp_service.GetSquareToolListReq{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	var list []response.ToolSquareInfo
	for _, item := range resp.Infos {
		list = append(list, toToolSquareInfo(ctx, item))
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

func UpsertBuiltinToolAPIKey(ctx *gin.Context, userID, orgID string, req request.ToolSquareAPIKeyReq) error {
	_, err := mcp.UpsertBuiltinToolAPIKey(ctx.Request.Context(), &mcp_service.UpsertBuiltinToolAPIKeyReq{
		ToolSquareId: req.ToolSquareID,
		ApiKey:       req.APIKey,
		Identity: &mcp_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// --- internal ---

func toToolSquareDetail(ctx *gin.Context, toolSquare *mcp_service.SquareToolDetail) *response.ToolSquareDetail {
	ret := &response.ToolSquareDetail{
		ToolSquareInfo: toToolSquareInfo(ctx, toolSquare.Info),
		ToolSquareActions: response.ToolSquareActions{
			NeedApiKeyInput: toolSquare.BuiltInTools.NeedApiKeyInput,
			APIKey:          toolSquare.BuiltInTools.ApiAuth.ApiKeyValue,
			Detail:          toolSquare.BuiltInTools.Detail,
			ActionSum:       int64(toolSquare.BuiltInTools.ActionSum),
			ApiAuth: util.ApiAuthWebRequest{
				AuthType:           toolSquare.BuiltInTools.ApiAuth.AuthType,
				ApiKeyHeaderPrefix: toolSquare.BuiltInTools.ApiAuth.ApiKeyHeaderPrefix,
				ApiKeyHeader:       toolSquare.BuiltInTools.ApiAuth.ApiKeyHeader,
				ApiKeyQueryParam:   toolSquare.BuiltInTools.ApiAuth.ApiKeyQueryParam,
				ApiKeyValue:        toolSquare.BuiltInTools.ApiAuth.ApiKeyValue,
			},
		},
		Schema: toolSquare.Schema,
	}
	for _, tool := range toolSquare.BuiltInTools.Tools {
		ret.Tools = append(ret.Tools, toToolAction(tool))
	}
	return ret
}

func toToolSquareInfo(ctx *gin.Context, toolSquareInfo *mcp_service.ToolSquareInfo) response.ToolSquareInfo {
	return response.ToolSquareInfo{
		ToolSquareID: toolSquareInfo.ToolSquareId,
		Avatar:       cacheToolAvatar(ctx, constant.ToolTypeBuiltIn, toolSquareInfo.AvatarPath),
		Name:         toolSquareInfo.Name,
		Desc:         toolSquareInfo.Desc,
		Tags:         getToolTags(toolSquareInfo.Tags),
	}
}

func getToolTags(tagString string) []string {
	if tagString == "" {
		return []string{}
	}
	return strings.Split(tagString, ",")
}
