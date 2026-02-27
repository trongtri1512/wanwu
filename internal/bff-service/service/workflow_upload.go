package service

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	net_url "net/url"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func FileUrlConvertBase64(ctx *gin.Context, req *request.FileUrlConvertBase64Req) (string, error) {
	resp, err := http.Get(req.FileUrl)
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_http_get", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_http_get", fmt.Sprintf("StatusCode: %d", resp.StatusCode))
	}
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_read", err.Error())
	}

	base64Str, base64StrWithPrefix, err := util.FileData2Base64(fileData, req.CustomPrefix)
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_convert_base64", err.Error())
	}
	if req.AddPrefix {
		return base64StrWithPrefix, nil
	} else {
		return base64Str, nil
	}
}

func UploadFileToWorkflow(ctx *gin.Context, req *request.WorkflowUploadFileReq) (*response.UploadFileByWorkflowResp, error) {
	file, err := req.File.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	base64Str := base64.StdEncoding.EncodeToString(fileBytes)
	return UploadFileByWorkflow(ctx, req.File.Filename, base64Str)
}

func UploadFileBase64ToWorkflow(ctx *gin.Context, req *request.WorkflowUploadFileByBase64Req) (*response.UploadFileByWorkflowResp, error) {
	var finalFileName string
	if req.FileName == "" {
		finalFileName = util.GenUUID()
	} else {
		finalFileName = req.FileName
	}

	base64Data := req.File
	var inferredExt string

	// 尝试从标准 Data URL 提取 MIME 类型
	if strings.HasPrefix(base64Data, "data:") && strings.Contains(base64Data, ";base64,") {
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) == 2 {
			header := parts[0]
			base64Data = parts[1]

			mimeType := strings.TrimPrefix(header, "data:")
			if idx := strings.Index(mimeType, ";"); idx != -1 {
				mimeType = mimeType[:idx]
			}

			if mimeType != "" {
				if exts, _ := mime.ExtensionsByType(mimeType); len(exts) > 0 {
					inferredExt = exts[0]
				}
			}
		}
	}

	var finalExt string
	if req.FileExt != "" {
		ext := strings.TrimPrefix(req.FileExt, ".")
		finalExt = "." + ext
	} else if inferredExt != "" {
		finalExt = inferredExt
	}

	if finalExt != "" {
		finalFileName += finalExt
	}

	return UploadFileByWorkflow(ctx, finalFileName, base64Data)
}

func UploadFileByWorkflow(ctx *gin.Context, fileName, file string) (*response.UploadFileByWorkflowResp, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.UploadFileUri)
	ret := &response.UploadFileByWorkflowResp{}
	requestBody := map[string]string{
		"name": fileName,
		"data": file,
	}
	if resp, err := resty.New().
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(requestBody).
		SetResult(ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_upload_file", err.Error())
	} else if resp.StatusCode() >= 300 {
		b, err := io.ReadAll(resp.RawResponse.Body)
		if err != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_upload_file", fmt.Sprintf("[%v] %v", resp.StatusCode(), err))
		}
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_upload_file", fmt.Sprintf("[%v] %v", resp.StatusCode(), string(b)))
	}
	return ret, nil
}
