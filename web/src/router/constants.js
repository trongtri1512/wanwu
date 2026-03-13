export const PERMS = {
  PERMISSION: 'permission', // 权限管理
  PERMISSION_USER: 'permission.user', // 权限管理-用户管理
  PERMISSION_ORG: 'permission.org', // 权限管理-组织管理
  PERMISSION_ROLE: 'permission.role', // 权限管理-角色管理
  SETTING: 'setting', // 平台配置

  MODEL_SERVICE: 'model', // 模型服务
  MODEL_MANAGE: 'model.model_management', // 模型服务-模型管理

  RESOURCE: 'resource', // 资源库
  KNOWLEDGE: 'resource.knowledge', // 资源库-知识库
  MCP_SERVICE: 'resource.mcp', // 资源库-MCP服务
  TOOL: 'resource.tool', // 资源库-工具
  PROMPT: 'resource.prompt', // 资源库-提示词
  SKILL: 'resource.skill', // 资源库-Skill
  SAFETY: 'resource.safety', // 资源库-安全护栏

  APP_SPACE: 'app', // 应用开发
  RAG: 'app.rag', // 应用开发-文本问答
  WORKFLOW: 'app.workflow', // 应用开发-工作流
  AGENT: 'app.agent', // 应用开发-智能体

  SQUARE: 'exploration', // 探索广场
  EXPLORE: 'exploration.app', // 探索广场-应用广场
  MCP: 'exploration.mcp', // 探索广场-MCP广场
  TEMPLATE: 'exploration.template', // 探索广场-模板广场

  OPERATION: 'operation', // 运营管理
  STATISTIC: 'operation.statistic_client', // 运营管理-统计分析
  OAUTH: 'operation.oauth', // 运营管理-OAuth密钥管理

  APP_OBSERVATION: 'app_observability', // 应用观测
  OBSERVATION_STATISTIC: 'app_observability.statistic', // 应用观测-统计看板

  API_KEY: 'api_key', // API Key管理
  API_KEY_MANAGE: 'api_key.api_key_management', // API Key管理-API Key管理
};
