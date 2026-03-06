package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/util"
)

const (
	builtinSkillsConfigDir = "configs/microservice/bff-service/configs/agent-skills"
)

type SkillsConfig struct {
	SkillId       string `json:"skillId" mapstructure:"skillId"`
	Name          string `json:"name" mapstructure:"name"`
	Avatar        string `json:"avatar" mapstructure:"avatar"`
	Author        string `json:"author" mapstructure:"author"`
	Desc          string `json:"desc" mapstructure:"desc"`
	MdPath        string `json:"mdPath" mapstructure:"mdPath"`
	SkillMarkdown []byte `json:"-" mapstructure:"-"`
}

type SkillCreatorConfig struct {
	Instruction    string        `json:"instruction" mapstructure:"instruction"`
	EnableThinking bool          `json:"enable_thinking" mapstructure:"enable_thinking"`
	Skills         []SkillConfig `json:"skills" mapstructure:"skills"`
}

type SkillConfig struct {
	Dir string `json:"dir" mapstructure:"dir"`
}

func (stf *SkillsConfig) AgentSkillZipToBytes(skillsId string) ([]byte, error) {
	return util.ZipDir(filepath.Join(builtinSkillsConfigDir, skillsId))
}

// FixFrontMatterFormat 确保 front matter 格式正确（配合前端正确渲染）
func FixFrontMatterFormat(content string) string {
	// 如果内容不以 --- 开头，直接返回
	if !strings.HasPrefix(content, "---") {
		return content
	}

	// 找到第一个 --- 和第二个 ---(中间包裹的为name、description等字段)
	firstEnd := 3 // 第一个 --- 占3个字符
	secondStart := strings.Index(content[firstEnd:], "---")
	if secondStart == -1 {
		return content // 没有结束标记
	}

	secondStart += firstEnd
	secondEnd := secondStart + 3

	// 处理字段，确保每个字段后面都有换行
	lines := strings.Split(strings.TrimSpace(content[firstEnd:secondStart]), "\n")
	var processedLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			processedLines = append(processedLines, line+"\n\n")
		}
	}
	// 重新构建内容
	result := "---\n\n" + strings.Join(processedLines, "") + "---" + content[secondEnd:]

	return result
}

// --- internal ---

func (stf *SkillsConfig) load() error {
	markdownPath := filepath.Join(builtinSkillsConfigDir, stf.MdPath)
	b, err := os.ReadFile(markdownPath)
	if err != nil {
		return fmt.Errorf("load skill %v markdown path %v err: %v", stf.SkillId, markdownPath, err)
	}

	// 处理 front matter 格式
	stf.SkillMarkdown = []byte(FixFrontMatterFormat(string(b)))
	return nil
}
