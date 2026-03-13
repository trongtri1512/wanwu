import request from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

/*---工作流模板---*/
export const getWorkflowTempList = data => {
  return request({
    url: `${USER_API}/workflow/template/list`,
    method: 'get',
    params: data,
  });
};
export const getWorkflowTempInfo = data => {
  return request({
    url: `${USER_API}/workflow/template/detail`,
    method: 'get',
    params: data,
  });
};
export const getWorkflowRecommendsList = data => {
  return request({
    url: `${USER_API}/workflow/template/recommend`,
    method: 'get',
    params: data,
  });
};
export const downloadWorkflow = params => {
  return request({
    url: `${USER_API}/workflow/template/download`,
    method: 'get',
    params,
    responseType: 'blob',
  });
};
export const copyWorkflowTemplate = data => {
  return request({
    url: `${USER_API}/workflow/template`,
    method: 'post',
    data,
  });
};

/*---提示词模板---*/
export const getPromptTempList = data => {
  return request({
    url: `${USER_API}/prompt/template/list`,
    method: 'get',
    params: data,
  });
};

export const copyPromptTemplate = data => {
  return request({
    url: `${USER_API}/prompt/template`,
    method: 'post',
    data,
  });
};

/*---自定义提示词---*/
export const getCustomPromptList = data => {
  return request({
    url: `${USER_API}/prompt/custom/list`,
    method: 'get',
    params: data,
  });
};

export const createCustomPrompt = data => {
  return request({
    url: `${USER_API}/prompt/custom`,
    method: 'post',
    data,
  });
};

export const editCustomPrompt = data => {
  return request({
    url: `${USER_API}/prompt/custom`,
    method: 'put',
    data,
  });
};

export const copyCustomPrompt = data => {
  return request({
    url: `${USER_API}/prompt/custom/copy`,
    method: 'post',
    data,
  });
};

export const deleteCustomPrompt = data => {
  return request({
    url: `${USER_API}/prompt/custom`,
    method: 'delete',
    data,
  });
};

/*---Skills---*/
export const getSkillTempList = data => {
  return request({
    url: `${USER_API}/agent/skill/list`,
    method: 'get',
    params: data,
  });
};
export const getSkillTempInfo = data => {
  return request({
    url: `${USER_API}/agent/skill/detail`,
    method: 'get',
    params: data,
  });
};
export const downloadSkill = params => {
  return request({
    url: `${USER_API}/agent/skill/download`,
    method: 'get',
    params,
    responseType: 'blob',
  });
};

// 获取自定义skills列表
export const getCustomSkillList = data => {
  return request({
    url: `${USER_API}/agent/skill/custom/list`,
    method: 'get',
    params: data,
  });
};

// 删除自定义skills
export const deleteCustomSkill = data => {
  return request({
    url: `${USER_API}/agent/skill/custom`,
    method: 'delete',
    data,
  });
};

// 查询自定义skills详情
export const getCustomSkillInfo = data => {
  return request({
    url: `${USER_API}/agent/skill/custom/detail`,
    method: 'get',
    params: data,
  });
};

// 创建自定义skills会话
export const createCustomSkillConversation = data => {
  return request({
    url: `${USER_API}/agent/skill/conversation`,
    method: 'post',
    data,
  });
};

// 删除自定义skill会话
export const delCustomSkillConversation = data => {
  return request({
    url: `${USER_API}/agent/skill/conversation`,
    method: 'delete',
    data,
  });
};

// 查询自定义skill会话列表
export const getCustomSkillConversationList = data => {
  return request({
    url: `${USER_API}/agent/skill/conversation/list`,
    method: 'get',
    params: data,
  });
};

// 查询自定义skill会话详情
export const getCustomSkillConversationDetail = data => {
  return request({
    url: `${USER_API}/agent/skill/conversation/detail`,
    method: 'get',
    params: data,
  });
};

// 自定义skill会话sse
export const getCustomSkillSSeUrl = () => {
  return `${USER_API}/agent/skill/conversation/chat`;
};

// 发送自定义skill到资源库
export const sendCustomSkillToResource = data => {
  return request({
    url: `${USER_API}/agent/skill/conversation/save`,
    method: 'post',
    data,
  });
};

// 创建自定义skills
export const createCustomSkill = data => {
  return request({
    url: `${USER_API}/agent/skill/custom`,
    method: 'post',
    data,
  });
};

// 校验自定义skills
export const checkCustomSkill = data => {
  return request({
    url: `${USER_API}/agent/skill/custom/check`,
    method: 'post',
    data,
  });
};

// 清空skillChat对话
export const clearSkillConversation = data => {
  return request({
    url: `${USER_API}/agent/skill/conversation/clear`,
    method: 'delete',
    data,
  });
};
