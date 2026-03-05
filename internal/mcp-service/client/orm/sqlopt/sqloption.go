package sqlopt

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type sqlOptions []SQLOption

func SQLOptions(opts ...SQLOption) sqlOptions {
	return opts
}

func (s sqlOptions) Apply(db *gorm.DB) *gorm.DB {
	for _, opt := range s {
		db = opt.Apply(db)
	}
	return db
}

type SQLOption interface {
	Apply(db *gorm.DB) *gorm.DB
}

type funcSQLOption func(db *gorm.DB) *gorm.DB

func (f funcSQLOption) Apply(db *gorm.DB) *gorm.DB {
	return f(db)
}

func WithID(id uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if id > 0 {
			return db.Where("id = ?", id)
		}
		return db
	})
}

func WithIDs(ids []uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	})
}

func WithOrgID(orgID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if orgID != "" {
			return db.Where("org_id = ?", orgID)
		}
		return db
	})
}

func WithUserID(userID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if userID != "" {
			return db.Where("user_id = ?", userID)
		}
		return db
	})
}

func WithMcpSquareId(mcpSquareId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if mcpSquareId != "" {
			return db.Where("mcp_square_id = ?", mcpSquareId)
		}
		return db
	})
}

func WithName(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if name != "" {
			return db.Where("name = ?", name)
		}
		return db
	})
}

func LikeName(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if name != "" {
			return db.Where("name LIKE ?", "%"+name+"%")
		}
		return db
	})
}

func WithFrom(from string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if from != "" {
			return db.Where("from = ?", from)
		}
		return db
	})
}

func WithUpdateLock() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Clauses(clause.Locking{
			Strength: "UPDATE",
		})
	})
}

func WithToolSquareID(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if id != "" {
			return db.Where("tool_square_id = ?", id)
		}
		return db
	})
}

func WithToolSquareIDList(idList []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(idList) > 0 {
			return db.Where("tool_square_id IN ?", idList)
		}
		return db
	})
}

func WithToolSquareIDEmpty() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("tool_square_id = '' or tool_square_id IS NULL")
	})
}

func WithToolSquareIDNotEmpty() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("tool_square_id != '' and tool_square_id IS NOT NULL")
	})
}

func WithMcpServerId(mcpServerId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if mcpServerId != "" {
			return db.Where("mcp_server_id = ?", mcpServerId)
		}
		return db
	})
}

func WithMcpServerToolId(mcpServerToolId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if mcpServerToolId != "" {
			return db.Where("mcp_server_tool_id = ?", mcpServerToolId)
		}
		return db
	})
}

func WithMcpServerIdList(mcpServerIdList []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("mcp_server_id IN ?", mcpServerIdList)
	})
}

func WithCustomSkillSaveIds(saveIdList []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(saveIdList) > 0 {
			return db.Where("save_id IN ?", saveIdList)
		}
		return db
	})
}
