package middleware

import (
	"errors"
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	KnowledgeView   int32 = 0
	KnowledgeEdit   int32 = 10
	KnowledgeGrant  int32 = 20
	KnowledgeSystem int32 = 30
)

// AuthKnowledgeDoc 校验知识库权限
func AuthKnowledgeDoc(fieldName string, permissionType int32) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()
		//1.获取value值
		value := getFieldValue(ctx, fieldName)
		if len(value) == 0 {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("docId is required"))
			ctx.Abort()
			return
		}
		//2.根据docId获取知识库id
		knowledgeId, err := searchKnowledgeId(ctx, value)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
		//3.校验用户授权权限
		err = knowledgeGrantUser(ctx, knowledgeId, permissionType)
		//4.异常处理
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
	}
}

// AuthKnowledgeQAPair 校验问答对
func AuthKnowledgeQAPair(fieldName string, permissionType int32) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()
		//1.获取value值
		value := getFieldValue(ctx, fieldName)
		if len(value) == 0 {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("qaPairId is required"))
			ctx.Abort()
			return
		}
		//2.根据QAPairId获取知识库id
		knowledgeId, err := searchKnowledgeIdByQAPairId(ctx, value)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
		//3.校验用户授权权限
		err = knowledgeGrantUser(ctx, knowledgeId, permissionType)
		//4.异常处理
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
	}
}

// AuthKnowledgeIfHas 校验知识库权限,允许字段为空
func AuthKnowledgeIfHas(fieldName string, permissionType int32) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()
		//1.获取value值
		value := getFieldValue(ctx, fieldName)
		if len(value) == 0 {
			ctx.Next()
			return
		}
		//2.校验用户授权权限
		err := knowledgeGrantUser(ctx, value, permissionType)
		//3.返回结果
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
	}
}

// AuthKnowledge 校验知识库权限
func AuthKnowledge(fieldName string, permissionType int32) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()
		//1.获取value值
		value := getFieldValue(ctx, fieldName)
		if len(value) == 0 {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("knowledgeId is required"))
			ctx.Abort()
			return
		}
		//2.校验用户授权权限
		err := knowledgeGrantUser(ctx, value, permissionType)
		//3.返回结果
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
	}
}

func searchKnowledgeId(ctx *gin.Context, docId string) (string, error) {
	docInfo, err := service.GetDocDetail(ctx, "", "", docId)
	if err != nil {
		return "", err
	}
	return docInfo.KnowledgeId, nil
}

func searchKnowledgeIdByQAPairId(ctx *gin.Context, qaPairId string) (string, error) {
	qaPairInfo, err := service.GetQAPairDetail(ctx, "", "", qaPairId)
	if err != nil {
		return "", err
	}
	return qaPairInfo.KnowledgeId, nil
}

func knowledgeGrantUser(ctx *gin.Context, knowledgeId string, permissionType int32) error {
	// userID
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	// orgID
	orgID, err := getOrgID(ctx)
	if err != nil {
		return err
	}

	// check user knowledge permission
	if err = service.CheckKnowledgeUserPermission(ctx, userID, orgID, knowledgeId, permissionType); err != nil {
		return err
	}
	return nil
}
