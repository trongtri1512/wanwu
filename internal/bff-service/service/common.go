package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/pkg/imaging"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

var (
	avatarCacheMu       sync.Mutex
	avatarCacheLocalDir = "cache"

	mcpAvatarCacheLocalDir = "cache/mcp"
)

func GetUserPermission(ctx *gin.Context, userID, orgID string) (*response.UserPermission, error) {
	resp, err := iam.GetUserPermission(ctx.Request.Context(), &iam_service.GetUserPermissionReq{
		UserId: userID,
		OrgId:  orgID,
	})
	if err != nil {
		return nil, err
	}
	user, err := iam.GetUserInfo(ctx.Request.Context(), &iam_service.GetUserInfoReq{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	return &response.UserPermission{
		OrgPermission:    toOrgPermission(ctx, resp),
		Language:         getLanguageByCode(user.Language),
		IsUpdatePassword: resp.LastUpdatePasswordAt != 0,
		Avatar:           cacheUserAvatar(ctx, user.AvatarPath),
	}, nil
}

func GetOrgSelect(ctx *gin.Context, userID string) (*response.Select, error) {
	resp, err := iam.GetOrgSelect(ctx.Request.Context(), &iam_service.GetOrgSelectReq{UserId: userID})
	if err != nil {
		return nil, err
	}
	return &response.Select{
		Select: toOrgIDNames(ctx, resp.Selects, userID == config.SystemAdminUserID),
	}, nil
}

// UploadAvatar 返回avatar在minio的objectPath
func UploadAvatar(ctx *gin.Context, fileHeader *multipart.FileHeader) (string, error) {
	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png":
	default:
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_type_error")
	}

	// 读取文件内容
	file, err := fileHeader.Open()
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_upload_error", err.Error())
	}
	defer func() { _ = file.Close() }()

	// 读取图片到内存缓冲区
	imgBuf := new(bytes.Buffer)
	if _, err := io.Copy(imgBuf, file); err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_upload_error", err.Error())
	}
	fileName := fmt.Sprintf("%s%s", util.GenUUID(), ext)
	// 生成存储路径，avatar/fileName前两位字母/fileName
	objectName := path.Join("avatar", fileName[:2], fileName)
	objectPath := path.Join(minio.BucketCustom, objectName)

	if _, err = minio.Custom().PutObject(ctx.Request.Context(), minio.BucketCustom, objectName, imgBuf.Bytes()); err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_upload_error", err.Error())
	}
	return objectPath, nil
}

// CacheAvatar 将avatar在minio的objectPath转为前端可访问的地址，同时在本地缓存avatar
// 例如 custom-upload/avatar/abc/def.png => /v1/static/avatar/abc/def.png
func CacheAvatar(ctx *gin.Context, avatarObjectPath string, isResize bool) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		return avatar
	}
	avatarCacheMu.Lock()
	defer avatarCacheMu.Unlock()

	avatar.Key = avatarObjectPath

	parts := strings.SplitN(avatarObjectPath, "/", 2)
	if len(parts) <= 1 {
		log.Errorf("cache avatar %v err: invalid objectPath", avatarObjectPath)
		return avatar
	}
	bucketName := parts[0]
	objectName := parts[1]
	filePath := filepath.Join(avatarCacheLocalDir, objectName)

	_, err := os.Stat(filePath)
	// 1 文件存在
	if err == nil {
		avatar.Path = filepath.Join("/v1", filePath)
		return avatar
	}
	// 2 系统错误
	if !os.IsNotExist(err) {
		log.Errorf("cache avatar %v check %v exist err: %v", avatarObjectPath, filePath, err)
		return avatar
	}
	// 3 文件不存在
	// 3.1 创建目录
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Errorf("cache avatar %v mkdir %v err: %v", avatarObjectPath, filepath.Dir(filePath), err)
		return avatar
	}
	// 3.2 下载文件
	b, err := minio.Custom().GetObject(ctx.Request.Context(), bucketName, objectName)
	if err != nil {
		log.Errorf("cache avatar %v minio download err: %v", avatarObjectPath, err)
		return avatar
	}
	// 3.3 压缩图像
	if isResize {
		compressedData, err := resizeImage(b)
		if err != nil {
			log.Warnf("cache avatar %v compress failed, using original: %v", avatarObjectPath, err)
			compressedData = b
		}
		// 3.3.1 写入压缩文件
		if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
			log.Errorf("cache avatar %v write file %v err: %v", avatarObjectPath, filePath, err)
			return avatar
		}
		avatar.Path = filepath.Join("/v1", filePath)
		return avatar
	}
	// 3.4 写入原文件
	if err := os.WriteFile(filePath, b, 0644); err != nil {
		log.Errorf("cache avatar %v write file %v err: %v", avatarObjectPath, filePath, err)
		return avatar
	}
	avatar.Path = filepath.Join("/v1", filePath)
	return avatar
}

func cacheAppAvatar(ctx *gin.Context, avatarObjectPath, appType string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" && appType == constant.AppTypeRag {
		avatar.Path = config.Cfg().DefaultIcon.RagIcon
		return avatar
	}
	if avatarObjectPath == "" && appType == constant.AppTypeAgent {
		avatar.Path = config.Cfg().DefaultIcon.AgentIcon
		return avatar
	}
	return CacheAvatar(ctx, avatarObjectPath, true)
}

func cacheUserAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.UserIcon
		return avatar
	}
	return CacheAvatar(ctx, avatarObjectPath, true)
}

// tool builtin & custom
func cacheToolAvatar(ctx *gin.Context, toolType string, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	switch toolType {
	case constant.ToolTypeCustom:
		if avatarObjectPath == "" {
			avatar.Path = config.Cfg().DefaultIcon.ToolIcon
			return avatar
		}
		return CacheAvatar(ctx, avatarObjectPath, true)
	case constant.ToolTypeBuiltIn:
		return cacheMCPServiceAvatar(ctx, avatarObjectPath)
	}
	return avatar
}

// skill custom
func cacheSkillAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	if avatarObjectPath == "" {
		return request.Avatar{Path: config.Cfg().DefaultIcon.CustomSkillIcon}
	}
	return CacheAvatar(ctx, avatarObjectPath, true)
}

// mcp square & custom
func cacheMCPAvatar(ctx *gin.Context, squareObjectPath, customObjectPath string) request.Avatar {
	if squareObjectPath == "" {
		avatar := request.Avatar{}
		if customObjectPath == "" {
			avatar.Path = config.Cfg().DefaultIcon.McpCustomIcon
			return avatar
		}
		return CacheAvatar(ctx, customObjectPath, true)
	}
	return cacheMCPServiceAvatar(ctx, squareObjectPath)
}

// mcp server
func cacheMCPServerAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.McpServerIcon
		return avatar
	}
	return CacheAvatar(ctx, avatarObjectPath, true)
}

// 用于缓存 内置工具、MCP广场 的图片（来源于mcp-service）
func cacheMCPServiceAvatar(ctx *gin.Context, avatarPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarPath == "" {
		return avatar
	}
	avatarCacheMu.Lock()
	defer avatarCacheMu.Unlock()

	filePath := filepath.Join(mcpAvatarCacheLocalDir, avatarPath)

	_, err := os.Stat(filePath)
	// 1 文件存在
	if err == nil {
		avatar.Path = filepath.Join("/v1", filePath)
		return avatar
	}
	// 2 系统错误
	if !os.IsNotExist(err) {
		log.Errorf("cache mcp avatar %v check %v exist err: %v", avatarPath, filePath, err)
		return avatar
	}
	// 3. 文件不存在
	// 3.1 创建目录
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Errorf("cache mcp avatar %v mkdir %v err: %v", avatarPath, filepath.Dir(filePath))
		return avatar
	}
	// 3.2 下载文件
	resp, err := mcp.GetMCPAvatar(ctx.Request.Context(), &mcp_service.GetMCPAvatarReq{AvatarPath: avatarPath})
	if err != nil {
		log.Errorf("cache mcp avatar %v download err: %v", avatarPath, err)
		return avatar
	}
	// 3.3 写入文件
	if err := os.WriteFile(filePath, resp.Data, 0644); err != nil {
		log.Errorf("cache mcp avatar %v write file %v err: %v", avatarPath, filePath, err)
		return avatar
	}
	avatar.Path = filepath.Join("/v1", filePath)
	return avatar
}

// cacheWorkflowAvatar 将avatar http请求地址转为前端统一访问的格式，同时在本地缓存avatar
// 例如 http://IP:port/api/static/abc/def.jpg => /v1/static/avatar/abc/def.png
func cacheWorkflowAvatar(avatarURL, appType string) request.Avatar {
	avatar := request.Avatar{}
	switch appType {
	case constant.AppTypeWorkflow:
		if avatarURL == "" {
			avatar.Path = config.Cfg().DefaultIcon.WorkflowIcon
			return avatar
		}
	case constant.AppTypeChatflow:
		if avatarURL == "" {
			avatar.Path = config.Cfg().DefaultIcon.ChatflowIcon
			return avatar
		}
	}

	avatarCacheMu.Lock()
	defer avatarCacheMu.Unlock()

	avatar.Key = avatarURL

	// 提取文件名：先去掉查询参数，再取最后一部分
	baseURL := avatarURL
	if idx := strings.Index(avatarURL, "?"); idx != -1 {
		baseURL = avatarURL[:idx]
	}
	// 从路径中提取文件名
	lastSlash := strings.LastIndex(baseURL, "/")
	fileName := baseURL[lastSlash+1:]
	filePath := filepath.Join(avatarCacheLocalDir, fileName)
	// 检查文件是否已缓存
	if _, err := os.Stat(filePath); err == nil {
		avatar.Path = filepath.Join("/v1", filePath)
		return avatar
	}
	var newAvatarURL string
	if strings.Contains(avatarURL, config.Cfg().Workflow.MinioProxyPrefix) {
		// 解析原始URL
		parsedURL, err := url.Parse(avatarURL)
		if err != nil {
			log.Errorf("cache avatar parse URL %v failed: %v", avatarURL, err)
			return avatar
		}
		// 去掉 /workflow/minio/presign/ 前缀
		path := parsedURL.Path
		path = strings.TrimPrefix(path, config.Cfg().Workflow.MinioProxyPrefix)
		// 使用 url.JoinPath 构建新URL
		newAvatarURL, err = url.JoinPath(config.Cfg().Workflow.MinioProxyEndpoint, path)
		if err != nil {
			log.Errorf("join path failed: %v", err)
			avatar.Path = avatarURL
			return avatar
		}
		// 添加查询参数
		if parsedURL.RawQuery != "" {
			newAvatarURL += "?" + parsedURL.RawQuery
		}
	} else if strings.Contains(avatarURL, config.Cfg().Server.WebBaseUrl) {
		// 直接转换URL为容器内可访问URL（如 http://localhost:8081/api/static/icon/icon-HTTP.png -> http://workflow-wanuw:8999/api/static/icon/icon-HTTP.png）
		parsedURL, err := url.Parse(avatarURL)
		if err != nil {
			log.Errorf("cache avatar parse URL %v failed: %v", avatarURL, err)
			return avatar
		}
		// 替换为workflow endpoint
		internalURL, err := url.Parse(config.Cfg().Workflow.Endpoint)
		if err != nil {
			log.Errorf("cache avatar invalid Workflow.Endpoint %s", config.Cfg().Workflow.Endpoint)
			return avatar
		}
		parsedURL.Host = internalURL.Host
		parsedURL.Scheme = internalURL.Scheme
		newAvatarURL = parsedURL.String()
	} else {
		newAvatarURL = avatarURL
	}
	// 从HTTP URL下载文件
	resp, err := http.Get(newAvatarURL)
	if err != nil {
		log.Errorf("cache avatar %v download %v err: %v", avatarURL, newAvatarURL, err)
		return avatar
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		log.Errorf("cache avatar %v download %v HTTP error: %v", avatarURL, newAvatarURL, resp.Status)
		return avatar
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("cache avatar %v download %v read response err: %v", avatarURL, newAvatarURL, err)
		return avatar
	}
	// 压缩图像
	compressedData, err := resizeImage(body)
	if err != nil {
		log.Warnf("cache avatar %v compress failed, using original: %v", avatarURL, err)
		// 压缩失败时使用原始数据
		compressedData = body
	}
	// 创建目录
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Errorf("cache avatar %v mkdir %v err: %v", avatarURL, filepath.Dir(filePath), err)
		return avatar
	}
	// 写入文件
	if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
		log.Errorf("cache avatar %v write file %v err: %v", avatarURL, filePath, err)
		return avatar
	}
	avatar.Path = filepath.Join("/v1", filePath)
	return avatar
}

func cachePromptAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.PromptIcon
		return avatar
	}
	return CacheAvatar(ctx, avatarObjectPath, true)
}

// knowledge custom
func cacheKnowledgeAvatar(ctx *gin.Context, avatarObjectPath string, knowledgeType int32) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		switch knowledgeType {
		case constant.KnowledgeBase:
			avatar.Path = config.Cfg().DefaultIcon.KnowledgeIcon
		case constant.KnowledgeQA:
			avatar.Path = config.Cfg().DefaultIcon.QAIcon
		default:
			avatar.Path = config.Cfg().DefaultIcon.KnowledgeIcon
		}
		return avatar
	}
	return CacheAvatar(ctx, avatarObjectPath, true)
}

// resizeImage 压缩图像
func resizeImage(imageData []byte) ([]byte, error) {
	// 先解码获取图像尺寸
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()
	// 计算等比例缩放后的尺寸
	targetWidth, targetHeight := calculateResizeParameters(originalWidth, originalHeight, 200)
	// 重新创建 reader（因为之前的读取位置已经改变）
	reader := bytes.NewReader(imageData)
	// 压缩图像到计算后的尺寸
	compressedData, err := imaging.Resize(reader, targetWidth, targetHeight)
	if err != nil {
		return nil, fmt.Errorf("image resize failed: %w", err)
	}
	return compressedData, nil
}

// 计算等比例缩放尺寸
func calculateResizeParameters(originalWidth, originalHeight, maxSize int) (int, int) {
	if originalWidth <= maxSize && originalHeight <= maxSize {
		// 如果原图已经小于目标尺寸，返回原尺寸
		return originalWidth, originalHeight
	}
	var newWidth, newHeight int
	if originalWidth > originalHeight {
		// 宽图：以宽度为基准
		newWidth = maxSize
		newHeight = int(float64(originalHeight) * float64(maxSize) / float64(originalWidth))
	} else {
		// 高图或正方形：以高度为基准
		newHeight = maxSize
		newWidth = int(float64(originalWidth) * float64(maxSize) / float64(originalHeight))
	}
	// 确保最小尺寸为1
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}
	return newWidth, newHeight
}

func convertStatisticChart(ctx *gin.Context, pbChart *common.StatisticChart) response.StatisticChart {
	if pbChart == nil {
		return response.StatisticChart{}
	}
	respChart := response.StatisticChart{
		TableName: gin_util.I18nKey(ctx, pbChart.TableName),
		Lines:     make([]response.StatisticChartLine, 0, len(pbChart.ChartLines)),
	}
	for _, pbLine := range pbChart.ChartLines {
		goLine := response.StatisticChartLine{
			LineName: gin_util.I18nKey(ctx, pbLine.LineName),
			Items:    make([]response.StatisticChartLineItem, 0, len(pbLine.Items)),
		}

		for _, pbItem := range pbLine.Items {
			goLine.Items = append(goLine.Items, response.StatisticChartLineItem{
				Key:   pbItem.Key,
				Value: pbItem.Value,
			})
		}
		respChart.Lines = append(respChart.Lines, goLine)
	}
	return respChart
}

func writeSSE(ctx *gin.Context, resp *http.Response) error {
	// 设置 SSE 响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("X-Accel-Buffering", "no") // 针对 Nginx 代理

	// 使用固定缓冲区读取
	buffer := make([]byte, 8192) // 8KB 缓冲区
	reader := bufio.NewReader(resp.Body)

	for {
		select {
		case <-ctx.Done():
			// 客户端断开连接
			return errors.New("writeSSE: ctx canceled")
		default:
			n, err := reader.Read(buffer)

			if n > 0 {
				if _, err := ctx.Writer.Write(buffer[:n]); err != nil {
					// 客户端可能已断开
					log.Errorf("writeSSE write err: %v", err)
					return err
				}
				ctx.Writer.Flush()
			}

			if err != nil {
				if err == io.EOF {
					return nil // 正常结束
				}
				log.Errorf("writeSSE read err: %v", err)
				return err
			}
		}
	}
}
