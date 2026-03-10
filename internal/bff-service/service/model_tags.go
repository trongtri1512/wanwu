package service

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

func GetTagsByScopeType(scopeType string) []mp_common.Tag {
	var tags []mp_common.Tag
	switch scopeType {
	case config.ModelScopeTypePrivate:
		tags = append(tags, mp_common.Tag{
			Text: mp_common.TagScopeTypePrivate,
		})
	case config.ModelScopeTypePublic:
		tags = append(tags, mp_common.Tag{
			Text: mp_common.TagScopeTypePublic,
		})
	case config.ModelScopeTypeOrg:
		tags = append(tags, mp_common.Tag{
			Text: mp_common.TagScopeTypeOrg,
		})
	}
	return tags
}

func GetTagsByImportSource(importSource string) []mp_common.Tag {
	var tags []mp_common.Tag
	switch importSource {
	case config.ModelSourceBuiltin:
		tags = append(tags, mp_common.Tag{
			Text: mp_common.TagSourceTypeLocal,
		})
	}
	return tags
}
