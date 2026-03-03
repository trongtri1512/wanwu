package option

import (
	"context"
	"encoding/json"
	"fmt"

	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3"
)

// invokableToolImpl 实现 eino 的 InvokableTool 接口。
// 基于 OpenAPI schema 调用 HTTP API。
type invokableToolImpl struct {
	doc    *openapi3.T         // OpenAPI schema
	auth   *openapi3_util.Auth // API 认证配置
	schema *schema.ToolInfo    // 工具信息
}

func (tool *invokableToolImpl) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return tool.schema, nil
}

func (tool *invokableToolImpl) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// operation
	var operation *openapi3.Operation
	for _, pathItem := range tool.doc.Paths {
		for _, op := range pathItem.Operations() {
			if op.OperationID == tool.schema.Name {
				operation = op
				break
			}
			if operation != nil {
				break
			}
		}
	}
	if operation == nil {
		return "", fmt.Errorf("action (%v) not found", tool.schema.Name)
	}
	// params
	var params map[string]any
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("action (%v) unmarshal args (%v) err: %v", tool.schema.Name, argumentsInJSON, err)
	}
	headerParams := make(map[string]string)
	pathParams := make(map[string]interface{})
	queryParams := make(map[string]interface{})
	bodyParams := make(map[string]interface{})
	for key, value := range params {
		// parameters
		for _, param := range operation.Parameters {
			if param.Value == nil || param.Value.Name != key {
				continue
			}
			switch param.Value.In {
			case "path":
				pathParams[key] = value
			case "query":
				queryParams[key] = value
			case "header":
				if valueStr, ok := value.(string); ok {
					headerParams[key] = valueStr
				} else {
					b, _ := json.Marshal(value)
					headerParams[key] = string(b)
				}
			}
		}
		// request body
		if operation.RequestBody != nil && operation.RequestBody.Value != nil {
			for _, mediaType := range operation.RequestBody.Value.Content {
				if mediaType.Schema != nil && mediaType.Schema.Value != nil {
					for propName := range mediaType.Schema.Value.Properties {
						if propName == key {
							bodyParams[propName] = value
						}
					}
				}
			}
		}
	}
	// auth
	if tool.auth != nil && tool.auth.Type != "" && tool.auth.Type != "none" && tool.auth.Value != "" {
		switch tool.auth.In {
		case "header":
			headerParams[tool.auth.Name] = tool.auth.Value
		case "query":
			queryParams[tool.auth.Name] = tool.auth.Value
		}
	}
	// http client
	client := openapi3_util.NewClientByDoc(tool.doc)
	// do request
	resp, err := client.DoRequestByOperationID(ctx, tool.schema.Name, &openapi3_util.RequestParams{
		HeaderParams: headerParams,
		PathParams:   pathParams,
		QueryParams:  queryParams,
		BodyParams:   bodyParams,
	})
	if err != nil {
		return "", err
	}
	if respStr, ok := resp.(string); ok {
		return respStr, nil
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("action (%v) marshal resp err: %v", tool.schema.Name, err)
	}
	return string(b), err
}
