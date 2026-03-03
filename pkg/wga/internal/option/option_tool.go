package option

import (
	"fmt"

	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
)

func (options *Options) checkToolsCondition(toolCategories []*config.ToolCategory) ([]CheckToolCategory, error) {
	var rets []CheckToolCategory
	for _, toolCategory := range toolCategories {
		// category tools
		var retTools []CheckTool
		for _, toolCfg := range toolCategory.Tools {
			if !toolCfg.AuthRequired {
				// 无需配置的tool，不加入该category的检查结果
				continue
			}
			var toolMeet bool
			for _, toolOpt := range options.Tools {
				if toolOpt.Title != toolCfg.Doc.Info.Title {
					continue
				}
				if auth, err := toolOpt.APIAuth.ToOpenapiAuth(); err != nil || auth.Type == "none" {
					continue
				}
				toolMeet = true
				break
			}
			retTools = append(retTools, CheckTool{
				Title: toolCfg.Doc.Info.Title,
				Meet:  toolMeet,
			})
		}

		if len(retTools) == 0 {
			// 该category下没有需要配置的tool，则无需返回该category的检查结果
			continue
		}
		// category meet
		var categoryMeet bool
		switch toolCategory.Condition {
		case config.ToolCategoryConditionNone:
			categoryMeet = true
		case config.ToolCategoryConditionOptional:
			for _, retTool := range retTools {
				if retTool.Meet {
					categoryMeet = true
					break
				}
			}
		case config.ToolCategoryConditionRequired:
			categoryMeet = true
			for _, retTool := range retTools {
				if !retTool.Meet {
					categoryMeet = false
					break
				}
			}
		default:
			return nil, fmt.Errorf("tool category (%v) condition (%v) unknown", toolCategory.Category, toolCategory.Condition)
		}
		rets = append(rets, CheckToolCategory{
			Category:  string(toolCategory.Category),
			Condition: string(toolCategory.Condition),
			Meet:      categoryMeet,
			Tools:     retTools,
		})
	}
	return rets, nil
}

func (options *Options) ToToolsConfig(toolCategories []*config.ToolCategory) (adk.ToolsConfig, error) {
	ret := adk.ToolsConfig{
		ToolsNodeConfig: compose.ToolsNodeConfig{
			ExecuteSequentially: true,
		},
		ReturnDirectly: make(map[string]bool),
	}

	conditions, err := options.checkToolsCondition(toolCategories)
	if err != nil {
		return ret, err
	}
	for _, condition := range conditions {
		if !condition.Meet {
			return ret, fmt.Errorf("tool category (%v) condition (%v) not meet", condition.Category, condition.Condition)
		}
	}

	for _, toolCategory := range toolCategories {
		for _, toolCfg := range toolCategory.Tools {
			// auth
			var auth *openapi3_util.Auth
			if toolCfg.AuthRequired {
				var err error
				for _, toolOpt := range options.Tools {
					if toolOpt.Title == toolCfg.Doc.Info.Title && toolOpt.APIAuth != nil {
						auth, err = toolOpt.APIAuth.ToOpenapiAuth()
						if err != nil {
							return ret, fmt.Errorf("tool (%v) auth convert err: %v", toolOpt.Title, err)
						}
						break
					}
				}
				if auth == nil {
					// 当前tool需要auth，但未配置，跳过
					continue
				}
			}
			// operation
			for _, operation := range toolCfg.Operations {
				schema, err := openapi3_util.Doc2EinoTool(toolCfg.Doc, operation.OperationID)
				if err != nil {
					return ret, fmt.Errorf("tool (%v) operation (%v) convert err: %v", toolCfg.Doc.Info.Title, operation.OperationID, err)
				}
				ret.Tools = append(ret.Tools, &invokableToolImpl{
					doc:    toolCfg.Doc,
					auth:   auth,
					schema: schema,
				})
				if operation.ReturnDirectly {
					ret.ReturnDirectly[operation.OperationID] = true
				}
			}
		}
	}
	return ret, nil
}
