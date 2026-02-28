import { PERMS } from '@/router/permission';
import { i18n } from '@/lang';

/**
 *  index: 为唯一标识，children 下定义的 index 标准为： 父级 index-子级定义的唯一标识
 */
export const menuList = [
  {
    name: i18n.t('menu.modelService'),
    index: 'modelService',
    icon: 'modelService',
    perm: [PERMS.MODEL_SERVICE],
    children: [
      {
        name: i18n.t('menu.modelAccess'),
        index: 'modelService-modelAccess',
        path: '/modelAccess',
        perm: PERMS.MODEL_MANAGE,
      },
    ],
  },
  {
    name: i18n.t('menu.aiAssistant'),
    index: 'aiAssistant',
    icon: 'aiAssistant',
    path: '/aiAssistant',
  },
  {
    name: i18n.t('menu.resource'),
    index: 'resource',
    icon: 'resource',
    perm: [PERMS.RESOURCE],
    children: [
      {
        name: i18n.t('menu.knowledge'),
        index: 'resource-knowledge',
        path: '/knowledge',
        perm: PERMS.KNOWLEDGE,
      },
      {
        name: i18n.t('menu.mcpService'),
        index: 'resource-mcpService',
        path: '/mcpService',
        perm: PERMS.MCP_SERVICE,
      },
      {
        name: i18n.t('menu.tool'),
        index: 'resource-tool',
        path: '/tool',
        perm: PERMS.TOOL,
      },
      {
        name: i18n.t('menu.prompt'),
        index: 'resource-prompt',
        path: '/prompt',
        perm: PERMS.PROMPT,
      },
      {
        name: 'Skills',
        index: 'resource-skill',
        path: '/skill',
        perm: PERMS.SKILL,
      },
      {
        name: i18n.t('menu.safetyGuard'),
        index: 'resource-safetyGuard',
        path: '/safety',
        perm: PERMS.SAFETY,
      },
    ],
  },
  {
    name: i18n.t('menu.app.index'),
    index: 'appSpace',
    perm: [PERMS.APP_SPACE],
    icon: 'appSpace',
    children: [
      {
        name: i18n.t('menu.app.rag'),
        index: 'appSpace-rag',
        path: '/appSpace/rag',
        perm: PERMS.RAG,
      },
      {
        name: i18n.t('menu.app.workflow'),
        index: 'appSpace-workflow',
        path: '/appSpace/workflow',
        perm: PERMS.WORKFLOW,
      },
      {
        name: i18n.t('menu.app.agent'),
        index: 'appSpace-agent',
        path: '/appSpace/agent',
        perm: PERMS.AGENT,
      },
    ],
  },
  {
    name: i18n.t('menu.square'),
    index: 'square',
    perm: [PERMS.SQUARE],
    icon: 'square',
    children: [
      {
        name: i18n.t('menu.explore'),
        index: 'square-explore',
        path: '/explore',
        perm: PERMS.EXPLORE,
      },
      {
        name: i18n.t('menu.mcp'),
        index: 'square-mcpManage',
        path: '/mcp',
        perm: PERMS.MCP,
      },
      {
        name: i18n.t('menu.templateSquare'),
        index: 'square-templateSquare',
        path: '/templateSquare',
        perm: PERMS.TEMPLATE,
      },
    ],
  },
  {
    name: i18n.t('menu.apiKey'),
    index: 'apiKey',
    icon: 'apiKey',
    perm: [PERMS.API_KEY],
    children: [
      {
        name: i18n.t('menu.apiKey'),
        index: 'apiKey-openApiKey',
        path: '/openApiKey',
        perm: PERMS.API_KEY_MANAGE,
      },
    ],
  },
];
