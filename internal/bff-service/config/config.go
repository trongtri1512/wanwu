package config

import (
	"fmt"
	"net/url"

	oauth2_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/oauth2-util"
	"github.com/UnicomAI/wanwu/pkg/i18n"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
)

var (
	_c *Config
)

type Config struct {
	Server            ServerConfig               `json:"server" mapstructure:"server"`
	Log               LogConfig                  `json:"log" mapstructure:"log"`
	JWT               JWTConfig                  `json:"jwt" mapstructure:"jwt"`
	OAuth             OAuthConfig                `json:"oauth" mapstructure:"oauth"`
	Decrypt           DecryptPasswd              `json:"decrypt-passwd" mapstructure:"decrypt-passwd"`
	I18n              i18n.Config                `json:"i18n" mapstructure:"i18n"`
	CustomInfo        CustomInfoConfig           `json:"custom-info" mapstructure:"custom-info"`
	DocCenter         DocCenterConfig            `json:"doc-center" mapstructure:"doc-center"`
	DefaultIcon       DefaultIconConfig          `json:"default-icon" mapstructure:"default-icon"`
	AssistantTemplate AssistantTemplateConfig    `json:"assistant-template" mapstructure:"assistant-template"`
	WorkflowTemplate  WorkflowTemplatePathConfig `json:"workflow-template" mapstructure:"workflow-template"`
	PromptTemplate    PromptTemplatePathConfig   `json:"prompt-template" mapstructure:"prompt-template"`
	SkillsTemplate    SkillsTemplatePathConfig   `json:"skills-template" mapstructure:"skills-template"`
	WorkflowTemplates []*WorkflowTemplateConfig  `json:"workflows" mapstructure:"workflows"`
	PromptTemplates   []*PromptTempConfig        `json:"prompts" mapstructure:"prompts"`
	AgentSkills       []*SkillsConfig            `json:"skills" mapstructure:"skills"`
	PromptEngineering PromptEngineeringConfig    `json:"prompt-engineering" mapstructure:"prompt-engineering"`
	// middleware
	Minio minio.Config `json:"minio" mapstructure:"minio"`
	Redis redis.Config `json:"redis" mapstructure:"redis"`
	// microservice
	Iam                 ServiceConfig         `json:"iam" mapstructure:"iam"`
	Model               ServiceModelConfig    `json:"model" mapstructure:"model"`
	MCP                 ServiceConfig         `json:"mcp" mapstructure:"mcp"`
	App                 ServiceConfig         `json:"app" mapstructure:"app"`
	Knowledge           ServiceConfig         `json:"knowledge" mapstructure:"knowledge"`
	Rag                 ServiceConfig         `json:"rag" mapstructure:"rag"`
	Assistant           ServiceConfig         `json:"assistant" mapstructure:"assistant"`
	Operate             ServiceConfig         `json:"operate" mapstructure:"operate"`
	RagKnowledgeConfig  RagKnowledgeConfig    `json:"rag-knowledge" mapstructure:"rag-knowledge"`
	DifyKnowledgeConfig DifyKnowledgeConfig   `json:"dify-knowledge" mapstructure:"dify-knowledge"`
	Workflow            WorkflowServiceConfig `json:"workflow" mapstructure:"workflow"`
	WgaSandbox          WgaSandboxConfig      `json:"wga-sandbox" mapstructure:"wga-sandbox"`
}

type ServerConfig struct {
	Host        string `json:"host" mapstructure:"host"`
	Port        int    `json:"port" mapstructure:"port"`
	WebBaseUrl  string `json:"web_base_url" mapstructure:"web_base_url"`
	ApiBaseUrl  string `json:"api_base_url" mapstructure:"api_base_url"`
	AppOpenUrl  string `json:"app_open_base_url" mapstructure:"app_open_base_url"`
	CallbackUrl string `json:"callback_url" mapstructure:"callback_url"`
}

type ServiceModelConfig struct {
	Host            string `json:"host" mapstructure:"host"`
	PngTestFilePath string `json:"png_test_file_path" mapstructure:"png_test_file_path"`
	PdfTestFilePath string `json:"pdf_test_file_path" mapstructure:"pdf_test_file_path"`
	AsrTestFilePath string `json:"asr_test_file_path" mapstructure:"asr_test_file_path"`
}

type LogConfig struct {
	Std   bool         `json:"std" mapstructure:"std"`
	Level string       `json:"level" mapstructure:"level"`
	Logs  []log.Config `json:"logs" mapstructure:"logs"`
}

type JWTConfig struct {
	SigningKey string `json:"signing-key" mapstructure:"signing-key"`
}

type OAuthSwitch struct {
	OAuthSwitch int `json:"switch" mapstructure:"switch"`
}

type OAuthJWTConfig struct {
	RSAPrivateKeyPath string `json:"private_key_path" mapstructure:"private_key_path"`
	RSAPublicKeyPath  string `json:"public_key_path" mapstructure:"public_key_path"`
}

type OAuthConfig struct {
	Switch int                   `json:"switch" mapstructure:"switch"`
	RSA    oauth2_util.RSAConfig `json:"rsa" mapstructure:"rsa"`
}

type DecryptPasswd struct {
	IV  string `json:"iv" mapstructure:"iv"`
	Key string `json:"key" mapstructure:"key"`
}

type ServiceConfig struct {
	Host string `json:"host" mapstructure:"host"`
}

type RagKnowledgeConfig struct {
	Endpoint               string `json:"endpoint" mapstructure:"endpoint"`
	ChatEndpoint           string `json:"chat-endpoint" mapstructure:"chat-endpoint"`
	UploadEndpoint         string `json:"upload-endpoint" mapstructure:"upload-endpoint"`
	SearchKnowledgeBaseUri string `json:"search-knowledge-base-uri" mapstructure:"search-knowledge-base-uri"`
	SearchKnowTimeout      int    `json:"search-know-timeout" mapstructure:"search-know-timeout"`
	SearchQABaseUri        string `json:"search-qa-base-uri" mapstructure:"search-qa-base-uri"`
	KnowledgeChatUri       string `json:"knowledge-chat-uri" mapstructure:"knowledge-chat-uri"`
	UploadUri              string `json:"upload-uri" mapstructure:"upload-uri"`
	UploadBucket           string `json:"upload-bucket" mapstructure:"upload-bucket"`
	UploadTime             int64  `json:"upload-timeout" mapstructure:"upload-timeout"` //单位s
}

type DifyKnowledgeConfig struct {
	SearchKnowledgeBaseUri string `json:"search-knowledge-base-uri" mapstructure:"search-knowledge-base-uri"`
	SearchKnowTimeout      int    `json:"search-know-timeout" mapstructure:"search-know-timeout"`
}

type WgaSandboxConfig struct {
	Sandbox WgaSandboxSandboxConfig `json:"sandbox" mapstructure:"sandbox"`
}

type WgaSandboxSandboxConfig struct {
	Type      string `json:"type" mapstructure:"type"`
	Host      string `json:"host" mapstructure:"host"`
	ImageName string `json:"image-name" mapstructure:"image-name"`
}

type WorkflowTemplatePathConfig struct {
	ServerMode string `json:"server_mode" mapstructure:"server_mode"`
	ConfigPath string `json:"configPath" mapstructure:"configPath"`

	GlobalWebListUrl string `json:"global_web_list_url" mapstructure:"global_web_list_url"`

	ListUrl      string `json:"list_url" mapstructure:"list_url"`
	DownloadUrl  string `json:"download_url" mapstructure:"download_url"`
	DetailUrl    string `json:"detail_url" mapstructure:"detail_url"`
	RecommendUrl string `json:"recommend_url" mapstructure:"recommend_url"`
}

type PromptTemplatePathConfig struct {
	ConfigPath string `json:"configPath" mapstructure:"configPath"`
}

type SkillsTemplatePathConfig struct {
	ConfigPath string `json:"configPath" mapstructure:"configPath"`
}

type PromptTempConfig struct {
	TemplateId string `json:"templateId" mapstructure:"templateId"`
	Category   string `json:"category" mapstructure:"category"`
	Avatar     string `json:"avatar"`
	Name       string `json:"name"`
	Desc       string `json:"desc" mapstructure:"desc"`
	Author     string `json:"author" mapstructure:"author"`
	Prompt     string `json:"prompt" mapstructure:"prompt"`
}

type PromptEngineeringConfig struct {
	Optimization string `json:"optimization" mapstructure:"optimization"`
}

type WorkflowServiceConfig struct {
	Endpoint           string `json:"endpoint" mapstructure:"endpoint"`
	MinioProxyEndpoint string `json:"minio_proxy_endpoint" mapstructure:"minio_proxy_endpoint"`
	MinioProxyPrefix   string `json:"minio_proxy_prefix" mapstructure:"minio_proxy_prefix"`
	// general
	ListUri    string `json:"list_uri" mapstructure:"list_uri"`
	CreateUri  string `json:"create_uri" mapstructure:"create_uri"`
	DeleteUri  string `json:"delete_uri" mapstructure:"delete_uri"`
	CopyUri    string `json:"copy_uri" mapstructure:"copy_uri"`
	ExportUri  string `json:"export_uri" mapstructure:"export_uri"`
	ImportUri  string `json:"import_uri" mapstructure:"import_uri"`
	ConvertUri string `json:"convert_uri" mapstructure:"convert_uri"`
	// run
	WorkflowRunByOpenapiUri     string `json:"workflow_run_by_openapi_uri" mapstructure:"workflow_run_by_openapi_uri"`
	WorkflowRunLatestVersionUri string `json:"workflow_run_latest_version_uri" mapstructure:"workflow_run_latest_version_uri"`
	GetProcessUri               string `json:"get_process_uri" mapstructure:"get_process_uri"`
	ChatflowRunByOpenapiUri     string `json:"chatflow_run_by_openapi_uri" mapstructure:"chatflow_run_by_openapi_uri"`
	// conversation
	CreateChatflowConversationUri string `json:"create_chatflow_conversation_uri" mapstructure:"create_chatflow_conversation_uri"`
	GetConversationMessageListUri string `json:"get_conversation_message_list_uri" mapstructure:"get_conversation_message_list_uri"`
	GetDraftIntelligenceListUri   string `json:"get_draft_intelligence_list_uri" mapstructure:"get_draft_intelligence_list_uri"`
	GetDraftIntelligenceInfoUri   string `json:"get_draft_intelligence_info_uri" mapstructure:"get_draft_intelligence_info_uri"`
	DeleteConversationUri         string `json:"delete_conversation_uri" mapstructure:"delete_conversation_uri"`
	GetProjectConversationDef     string `json:"get_project_conversation_def" mapstructure:"get_project_conversation_def"`
	// upload
	UploadActionUri string `json:"upload_action_uri" mapstructure:"upload_action_uri"`
	UploadCommonUri string `json:"upload_common_uri" mapstructure:"upload_common_uri"`
	UploadFileUri   string `json:"upload_file_uri" mapstructure:"upload_file_uri"`
	SignImgUri      string `json:"sign_img_uri" mapstructure:"sign_img_uri"`
	// version
	PublishUri           string               `json:"publish_uri" mapstructure:"publish_uri"`
	VersionListUri       string               `json:"version_list_uri" mapstructure:"version_list_uri"`
	UpdateVersionDescUri string               `json:"update_version_desc_uri" mapstructure:"update_version_desc_uri"`
	RollbackUri          string               `json:"rollback_uri" mapstructure:"rollback_uri"`
	ModelParams          []WorkflowModelParam `json:"model_params" mapstructure:"model_params"`
}

type WorkflowModelParam struct {
	Name      string `json:"name" mapstructure:"name"`
	Desc      string `json:"desc" mapstructure:"desc"`
	Label     string `json:"label" mapstructure:"label"`
	Type      int    `json:"type" mapstructure:"type"`
	Precision int    `json:"precision" mapstructure:"precision"`
	Min       string `json:"min" mapstructure:"min"`
	Max       string `json:"max" mapstructure:"max"`

	ParamClass WorkflowModelParamClass      `json:"param_class" mapstructure:"param_class"`
	DefaultVal WorkflowModelParamDefaultVal `json:"default_val" mapstructure:"default_val"`
}

type WorkflowModelParamClass struct {
	ClassID int    `json:"class_id" mapstructure:"class_id"`
	Label   string `json:"label" mapstructure:"label"`
}

type WorkflowModelParamDefaultVal struct {
	Precise    string `json:"precise" mapstructure:"precise"`
	Balance    string `json:"balance" mapstructure:"balance"`
	Creative   string `json:"creative" mapstructure:"creative"`
	DefaultVal string `json:"default_val" mapstructure:"default_val"`
}

type AssistantTemplateConfig struct {
	ConfigPath string `json:"configPath" mapstructure:"configPath"`
}

type DocCenterConfig struct {
	FrontendPrefix string          `json:"frontend_prefix" mapstructure:"frontend_prefix"`
	Links          []DocLinkConfig `json:"links" mapstructure:"links"`
	docs           map[string]string
}

type DocLinkConfig struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type CustomInfoConfig struct {
	DefaultMode          string        `json:"default_mode" mapstructure:"default_mode"`
	Modes                []CustomTheme `json:"modes" mapstructure:"modes"`
	Version              string        `json:"version" mapstructure:"version"`
	RegisterByEmail      int           `json:"register_by_email" mapstructure:"register_by_email"`
	ResetPasswordByEmail int           `json:"reset_password_by_email" mapstructure:"reset_password_by_email"`
	LoginByEmail         int           `json:"login_by_email" mapstructure:"login_by_email"`
}

type CustomTheme struct {
	Mode  string      `json:"mode" mapstructure:"mode"`
	Login CustomLogin `json:"login" mapstructure:"login"`
	Home  CustomHome  `json:"home" mapstructure:"home"`
	Tab   CustomTab   `json:"tab" mapstructure:"tab"`
	About CustomAbout `json:"about" mapstructure:"about"`
}

type CustomLogin struct {
	BackgroundPath   string `json:"background_path" mapstructure:"background_path"`
	LogoPath         string `json:"logo_path" mapstructure:"logo_path"`
	LoginButtonColor string `json:"login_button_color" mapstructure:"login_button_color"`
	WelcomeText      string `json:"welcome_text" mapstructure:"welcome_text"`
	PlatformDesc     string `json:"platform_desc" mapstructure:"platform_desc"`
}

type CustomHome struct {
	LogoPath        string `json:"logo_path" mapstructure:"logo_path"`
	Title           string `json:"title" mapstructure:"title"`
	BackgroundColor string `json:"background_color" mapstructure:"background_color"`
}

type CustomTab struct {
	TabTitle    string `json:"title" mapstructure:"title"`
	TabLogoPath string `json:"logo_path" mapstructure:"logo_path"`
}

type CustomAbout struct {
	LogoPath  string `json:"logo_path" mapstructure:"logo_path"`
	Copyright string `json:"copyright" mapstructure:"copyright"`
}

type DefaultIconConfig struct {
	UserIcon      string `json:"user" mapstructure:"user"`
	RagIcon       string `json:"rag" mapstructure:"rag"`
	AgentIcon     string `json:"agent" mapstructure:"agent"`
	WorkflowIcon  string `json:"workflow" mapstructure:"workflow"`
	ChatflowIcon  string `json:"chatflow" mapstructure:"chatflow"`
	McpCustomIcon string `json:"mcpCustom" mapstructure:"mcpCustom"`
	McpServerIcon string `json:"mcpServer" mapstructure:"mcpServer"`
	ToolIcon      string `json:"tool" mapstructure:"tool"`
	PromptIcon    string `json:"prompt" mapstructure:"prompt"`
	SkillIcon     string `json:"skill" mapstructure:"skill"`
}

func LoadConfig(in string) error {
	_c = &Config{}
	if err := util.LoadConfig(in, _c); err != nil {
		return err
	}
	_c.DocCenter.docs = make(map[string]string)
	for _, link := range _c.DocCenter.Links {
		url, _ := url.JoinPath(_c.Server.WebBaseUrl, _c.DocCenter.FrontendPrefix, url.PathEscape(link.Val))
		_c.DocCenter.docs[link.Key] = url
	}
	// 加载工作流模板配置
	if err := util.LoadConfig(_c.WorkflowTemplate.ConfigPath, _c); err != nil {
		return fmt.Errorf("load workflow template config err: %v", err)
	}
	for _, wt := range _c.WorkflowTemplates {
		if err := wt.load(); err != nil {
			return err
		}
	}
	// 加载提示词模板配置
	promptIn := _c.PromptTemplate.ConfigPath
	if err := util.LoadConfig(promptIn, _c); err != nil {
		return fmt.Errorf("load prompt template config err: %v", err)
	}
	// 加载skills模板配置
	skillsIn := _c.SkillsTemplate.ConfigPath
	if err := util.LoadConfig(skillsIn, _c); err != nil {
		return fmt.Errorf("load skills template config err: %v", err)
	}
	for _, st := range _c.AgentSkills {
		if err := st.load(); err != nil {
			return err
		}
	}
	return nil
}

func Cfg() *Config {
	if _c == nil {
		log.Panicf("cfg nil")
	}
	return _c
}

func (c *Config) WorkflowTemp(templateId string) (WorkflowTemplateConfig, bool) {
	for _, wtf := range c.WorkflowTemplates {
		if wtf.TemplateId == templateId {
			return *wtf, true
		}
	}
	return WorkflowTemplateConfig{}, false
}

func (c *Config) PromptTemp(templateId string) (PromptTempConfig, bool) {
	for _, ptf := range c.PromptTemplates {
		if ptf.TemplateId == templateId {
			return *ptf, true
		}
	}
	return PromptTempConfig{}, false
}

func (c *Config) AgentSkill(skillId string) (SkillsConfig, bool) {
	for _, stf := range c.AgentSkills {
		if stf.SkillId == skillId {
			return *stf, true
		}
	}
	return SkillsConfig{}, false
}

// GetDocs 返回 docs 的深拷贝
func (d *DocCenterConfig) GetDocs() map[string]string {
	if d.docs == nil {
		return nil
	}
	// 深拷贝
	result := make(map[string]string, len(d.docs))
	for k, v := range d.docs {
		result[k] = v
	}
	return result
}
