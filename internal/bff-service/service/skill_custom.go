package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	minio_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/minio-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	customSkillFileType = ".zip"
)

func extractSkillMarkdownFromZip(zipData []byte) (string, string, string, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read zip file: %v", err)
	}

	var skillMdFile *zip.File
	var allFiles []string
	for _, file := range reader.File {
		allFiles = append(allFiles, file.Name)
		fileName := filepath.Base(file.Name)
		if fileName == "SKILL.md" {
			skillMdFile = file
			break
		}
	}

	if skillMdFile == nil {
		return "", "", "", fmt.Errorf("SKILL.md file not found in the zip archive. Files: %v", allFiles)
	}

	rc, err := skillMdFile.Open()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to open SKILL.md file: %v", err)
	}
	defer func() { _ = rc.Close() }()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read SKILL.md file: %v", err)
	}

	markdownContent := string(content)
	fm, _, err := util.ParseSkillFrontMatter(markdownContent)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse frontmatter: %v", err)
	}

	var name, desc string
	if fm != nil {
		name = fm.Name
		desc = fm.Description
	}

	return markdownContent, name, desc, nil
}

func CreateCustomSkill(ctx *gin.Context, userId, orgId string, req request.CreateCustomSkillReq) (*response.CustomSkillIDResp, error) {
	var objectPath string
	var markdownContent string
	skillName := req.Name
	skillDesc := req.Desc

	if req.ZipUrl != "" {
		// 下载文件
		data, err := minio_util.DownloadFileDirect(ctx.Request.Context(), req.ZipUrl)
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("download skill zip err: %v", err))
		}

		// 解压并查找SKILL.md文件，提取name和description
		mdContent, name, desc, err := extractSkillMarkdownFromZip(data)
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
		}
		markdownContent = mdContent

		// 如果从markdown中提取到了name和desc，使用这些值
		if skillName == "" {
			skillName = name
		}
		if skillDesc == "" {
			skillDesc = desc
		}

		fileName, _, err := minio.UploadFileCommon(ctx.Request.Context(), bytes.NewReader(data), customSkillFileType, -1, true)
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
		}
		// 构建完整的相对路径：file-upload/file-not-expire/xxx.zip
		objectPath = filepath.Join(minio.BucketFileUpload, minio.DirFileNotExpire, fileName)
	}

	createResp, err := mcp.CustomSkillCreate(ctx.Request.Context(), &mcp_service.CustomSkillCreateReq{
		Name:       skillName,
		Avatar:     req.Avatar.Key,
		Author:     req.Author,
		Desc:       skillDesc,
		ObjectPath: objectPath,
		Markdown:   markdownContent,
		SaveId:     req.SaveId,
		SourceType: req.SourceType,
		Identity:   &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}

	return &response.CustomSkillIDResp{SkillId: createResp.SkillId}, nil
}

func GetCustomSkill(ctx *gin.Context, userId, orgId, skillId string) (*response.CustomSkillDetail, error) {
	resp, err := mcp.CustomSkillGet(ctx.Request.Context(), &mcp_service.CustomSkillGetReq{
		SkillId:  skillId,
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}

	return &response.CustomSkillDetail{
		SkillDetail: response.SkillDetail{
			SkillId:       resp.SkillId,
			Name:          resp.Name,
			Avatar:        cacheSkillAvatar(ctx, resp.Avatar),
			Author:        resp.Author,
			Desc:          resp.Desc,
			SkillMarkdown: config.FixFrontMatterFormat(resp.Markdown),
		},
		ZipUrl: buildAccessFilePath(resp.ObjectPath),
	}, nil
}

func DeleteCustomSkill(ctx *gin.Context, skillId string) error {
	_, err := mcp.CustomSkillDelete(ctx.Request.Context(), &mcp_service.CustomSkillDeleteReq{
		SkillId: skillId,
	})
	return err
}

func GetCustomSkillList(ctx *gin.Context, userId, orgId, name string) (*response.ListResult, error) {
	resp, err := mcp.CustomSkillGetList(ctx.Request.Context(), &mcp_service.CustomSkillGetListReq{
		Name:     name,
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}

	customSkillList := make([]*response.CustomSkillDetail, 0, len(resp.List))
	for _, skill := range resp.List {
		customSkillList = append(customSkillList, toCustomSkill(ctx, skill))
	}

	return &response.ListResult{
		List:  customSkillList,
		Total: resp.Total,
	}, nil
}

func toCustomSkill(ctx *gin.Context, skill *mcp_service.CustomSkill) *response.CustomSkillDetail {
	if skill == nil {
		return nil
	}
	return &response.CustomSkillDetail{
		SkillDetail: response.SkillDetail{
			SkillId: skill.SkillId,
			Name:    skill.Name,
			Avatar:  cacheSkillAvatar(ctx, skill.Avatar),
			Author:  skill.Author,
			Desc:    skill.Desc,
		},
		ZipUrl: buildAccessFilePath(skill.ObjectPath),
	}
}
