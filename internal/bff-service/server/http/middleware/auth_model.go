package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// AuthModel 校验多个可能的 modelId 字段路径
// fieldPaths: 如 []string{"modelConfig.modelId", "recommendConfig.modelId"}
func AuthModel(fields []string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()

		var modelIds []string

		// 1. 尝试从 Query 参数获取（仅支持顶层字段，不支持嵌套）
		for _, field := range fields {
			// 如果路径不含 "."，说明可能是 query 参数
			if !strings.Contains(field, ".") {
				if val := ctx.Query(field); val != "" {
					modelIds = append(modelIds, val)
				}
			}
		}

		// 2. 从 JSON Body 提取（支持嵌套）
		if ctx.ContentType() == binding.MIMEJSON {
			bodyStr, _ := requestBody(ctx)
			if bodyStr != "" {
				var paramsMap map[string]interface{}
				if json.Unmarshal([]byte(bodyStr), &paramsMap) == nil {
					for _, field := range fields {
						if val, ok := getNestedValue(paramsMap, field); ok {
							modelID, ok := val.(string)
							if !ok {
								gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("field %s must be string type", field)))
								return
							}
							modelIds = append(modelIds, modelID)
						}
					}
				}
			}
		}

		// 3. 去重
		uniqueModelIds := make(map[string]bool)
		var finalModelIds []string
		for _, id := range modelIds {
			if !uniqueModelIds[id] {
				uniqueModelIds[id] = true
				finalModelIds = append(finalModelIds, id)
			}
		}

		// 4. 如果没有 modelId，直接放行
		if len(finalModelIds) == 0 {
			ctx.Next()
			return
		}

		// 5. 获取用户和组织 ID
		userID, _ := getUserID(ctx)
		if userID == "" {
			gin_util.ResponseErrCodeKeyWithStatus(ctx, http.StatusBadRequest, err_code.Code_BFFAuth, "", "auth model userID not found")
			ctx.Abort()
			return
		}

		orgID, _ := getOrgID(ctx)
		if orgID == "" {
			gin_util.ResponseErrCodeKeyWithStatus(ctx, http.StatusBadRequest, err_code.Code_BFFAuth, "", "auth model orgID not found")
			ctx.Abort()
			return
		}

		// 6. 逐个校验模型权限

		if _, err := service.CheckModelUserPermission(ctx, userID, orgID, finalModelIds); err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// getNestedValue 从 map[string]interface{} 中按 "a.b.c" 路径取值
func getNestedValue(data map[string]interface{}, path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	var current interface{} = data

	for _, key := range keys {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[key]
			if current == nil {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return current, true
}
