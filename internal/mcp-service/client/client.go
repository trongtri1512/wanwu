package client

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
)

type IClient interface {
	CheckMCPExist(ctx context.Context, orgID, userID, mcpSquareID string) (bool, *errs.Status)
	GetMCP(ctx context.Context, mcpID uint32) (*model.MCPClient, *errs.Status)
	CreateMCP(ctx context.Context, mcp *model.MCPClient) *errs.Status
	UpdateMCP(ctx context.Context, mcp *model.MCPClient) *errs.Status
	DeleteMCP(ctx context.Context, mcpID uint32) *errs.Status
	ListMCPs(ctx context.Context, orgID, userID, name string) ([]*model.MCPClient, *errs.Status)
	ListMCPsByMCPIdList(ctx context.Context, mcpIDList []uint32) ([]*model.MCPClient, *errs.Status)

	CreateCustomTool(ctx context.Context, customTool *model.CustomTool) *errs.Status
	GetCustomTool(ctx context.Context, customTool *model.CustomTool) (*model.CustomTool, *errs.Status)
	ListCustomTools(ctx context.Context, orgID, userID, name string) ([]*model.CustomTool, *errs.Status)
	ListCustomToolsByCustomToolIDs(ctx context.Context, customToolIDs []uint32) ([]*model.CustomTool, *errs.Status)
	UpdateCustomTool(ctx context.Context, customTool *model.CustomTool) *errs.Status
	DeleteCustomTool(ctx context.Context, customToolID uint32) *errs.Status
	ListBuiltinTools(ctx context.Context, orgID, userID string) ([]*model.BuiltinTool, *errs.Status)
	ListBuiltinToolsBySquareIdList(ctx context.Context, squareList []string) ([]*model.BuiltinTool, *errs.Status)
	GetBuiltinTool(ctx context.Context, builtinTool *model.BuiltinTool) (*model.BuiltinTool, *errs.Status)

	UpdateBuiltinTool(ctx context.Context, builtinTool *model.BuiltinTool) *errs.Status
	CreateBuiltinTool(ctx context.Context, customTool *model.BuiltinTool) *errs.Status

	GetMCPServer(ctx context.Context, mcpServerId string) (*model.MCPServer, *errs.Status)
	CreateMCPServer(ctx context.Context, mcpServer *model.MCPServer) *errs.Status
	UpdateMCPServer(ctx context.Context, mcpServer *model.MCPServer) *errs.Status
	DeleteMCPServer(ctx context.Context, mcpServerId string) *errs.Status
	ListMCPServers(ctx context.Context, orgID, userID, name string) ([]*model.MCPServer, *errs.Status)
	ListMCPServerByIdList(ctx context.Context, mcpServerIdList []string) ([]*model.MCPServer, *errs.Status)
	GetMCPServerTool(ctx context.Context, mcpServerToolId string) (*model.MCPServerTool, *errs.Status)
	CreateMCPServerTool(ctx context.Context, mcpServerTools []*model.MCPServerTool) *errs.Status
	UpdateMCPServerTool(ctx context.Context, mcpServerTool *model.MCPServerTool) *errs.Status
	DeleteMCPServerTool(ctx context.Context, mcpServerToolId string) *errs.Status
	ListMCPServerTools(ctx context.Context, mcpServerId string) ([]*model.MCPServerTool, *errs.Status)
	CountMCPServerTools(ctx context.Context, mcpServerId string) (int64, *errs.Status)

	//================CustomSkill================
	CreateCustomSkill(ctx context.Context, customSkill *model.CustomSkill) (string, *errs.Status)
	DeleteCustomSkill(ctx context.Context, skillId string) *errs.Status
	GetCustomSkill(ctx context.Context, skillId string) (*model.CustomSkill, *errs.Status)
	GetCustomSkillList(ctx context.Context, userId, orgId, name string) ([]*model.CustomSkill, int64, *errs.Status)
	GetCustomSkillBySaveIds(ctx context.Context, saveIds []string) ([]*model.CustomSkill, *errs.Status)
}
