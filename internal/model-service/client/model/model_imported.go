package model

type ModelImported struct {
	ID             uint32 `gorm:"primary_key;auto_increment;not null;"`
	UUID           string `gorm:"column:uuid;type:varchar(255);uniqueIndex:idx_unique_uuid;comment:模型uuid"`
	Provider       string `gorm:"column:provider;index:idx_model_imported_provider_type_model,priority:1;type:varchar(100);comment:模型供应商"`
	ModelType      string `gorm:"column:model_type;index:idx_model_imported_provider_type_model,priority:2;type:varchar(100);comment:模型类型"`
	Model          string `gorm:"column:model;index:idx_model_imported_provider_type_model,priority:3;type:varchar(100);comment:模型名称"`
	DisplayName    string `gorm:"column:display_name;idx:idx_model_imported_model_display_name;type:varchar(100);comment:模型显示名称"`
	ModelIconPath  string `gorm:"column:model_icon_path;type:varchar(512);comment:模型图标路径"`
	IsActive       bool   `gorm:"column:is_active;type:tinyint(1);default:true;comment:模型是否启用"`
	ProviderConfig string `gorm:"column:provider_config;type:longtext;comment:某供应商下的模型配置"`
	ModelDesc      string `gorm:"column:model_desc;type:longtext;comment:模型描述"`
	PublishDate    string `gorm:"column:publish_date;type:varchar(100);comment:模型发布时间"`
	ScopeType      uint32 `gorm:"column:scope_type;type:int(10) unsigned;default:1;comment:模型公开范围 1-私有 2-公开 3-当前组织可见"`
	ImportSource   string `gorm:"column:import_source;type:varchar(50);default:external;comment:模型导入来源(builtin=平台内置,external=外部URL)"`
	PublicModel
}
