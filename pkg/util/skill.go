package util

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type FrontMatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func ParseSkillFrontMatter(content string) (*FrontMatter, string, error) {
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "---") {
		return nil, content, nil
	}

	rest := content[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return nil, content, nil
	}

	frontMatterStr := strings.TrimSpace(rest[:endIdx])
	markdownContent := rest[endIdx+4:]
	markdownContent = strings.TrimPrefix(markdownContent, "\n")

	var fm FrontMatter
	if err := yaml.Unmarshal([]byte(frontMatterStr), &fm); err != nil {
		return nil, markdownContent, err
	}

	return &fm, markdownContent, nil
}
