import { i18n } from '@/lang';

export const LLM = 'llm';
export const RERANK = 'rerank';
export const EMBEDDING = 'embedding';
export const MULTIMODAL_RERANK = 'multimodal-rerank';
export const MULTIMODAL_EMBEDDING = 'multimodal-embedding';
export const OCR = 'ocr';
export const GUI = 'gui';
export const PDF_PARSER = 'pdf-parser';
export const ASR = 'sync-asr';

export const MODEL_TYPE_OBJ = {
  [LLM]: 'LLM',
  [RERANK]: 'Text Rerank',
  [EMBEDDING]: 'Text Embedding',
  [MULTIMODAL_RERANK]: 'Multimodal Rerank',
  [MULTIMODAL_EMBEDDING]: 'Multimodal Embedding',
  [OCR]: 'OCR',
  [GUI]: 'GUI',
  [PDF_PARSER]: i18n.t('modelAccess.type.pdfParser'),
  [ASR]: i18n.t('modelAccess.type.asr'),
};

export const MODEL_TYPE = Object.keys(MODEL_TYPE_OBJ).map(key => ({
  key,
  name: MODEL_TYPE_OBJ[key],
}));

export const YUAN_JING = 'YuanJing';
export const OPENAI_API = 'OpenAI-API-compatible';
export const OLLAMA = 'Ollama';
export const QWEN = 'Qwen';
export const HUOSHAN = 'HuoShan';
export const INFINI = 'Infini';
export const DEEPSEEK = 'DeepSeek';
export const QIANFAN = 'QianFan';
export const JINA = 'Jina';

export const PROVIDER_OBJ = {
  [OPENAI_API]: 'OpenAI-API-compatible',
  [YUAN_JING]: i18n.t('modelAccess.type.yuanjing'),
  [OLLAMA]: 'Ollama',
  [QWEN]: i18n.t('modelAccess.type.qwen'),
  [HUOSHAN]: i18n.t('modelAccess.type.huoshan'),
  [INFINI]: i18n.t('modelAccess.type.infini'),
  [DEEPSEEK]: 'DeepSeek',
  [QIANFAN]: i18n.t('modelAccess.type.qianfan'),
  [JINA]: 'Jina',
};

export const PROVIDER_IMG_OBJ = {
  [OPENAI_API]: require('@/assets/imgs/openAI.png'),
  [YUAN_JING]: require('@/assets/imgs/yuanjing.png'),
  [OLLAMA]: require('@/assets/imgs/ollama.png'),
  [QWEN]: require('@/assets/imgs/qwen.png'),
  [HUOSHAN]: require('@/assets/imgs/volcano.png'),
  [INFINI]: require('@/assets/imgs/infini.png'),
  [DEEPSEEK]: require('@/assets/imgs/deepseek.png'),
  [QIANFAN]: require('@/assets/imgs/qianfan.png'),
  [JINA]: require('@/assets/imgs/jina.png'),
};

const COMMON_MODEL_KEY = [LLM, RERANK, EMBEDDING];
const OLL_MODEL_KEY = [LLM, EMBEDDING];
const MULTIMODAL_KEY = [MULTIMODAL_RERANK, MULTIMODAL_EMBEDDING];
export const PROVIDER_MODEL_KEY = {
  [OPENAI_API]: COMMON_MODEL_KEY,
  [YUAN_JING]: [
    ...COMMON_MODEL_KEY,
    MULTIMODAL_RERANK,
    OCR,
    GUI,
    PDF_PARSER,
    ASR,
  ], // ...MULTIMODAL_KEY
  [OLLAMA]: OLL_MODEL_KEY,
  [QWEN]: [...COMMON_MODEL_KEY, ASR],
  [HUOSHAN]: [...OLL_MODEL_KEY, ASR],
  [INFINI]: COMMON_MODEL_KEY,
  [DEEPSEEK]: [LLM],
  [QIANFAN]: COMMON_MODEL_KEY,
  [JINA]: [RERANK, EMBEDDING, ...MULTIMODAL_KEY],
};

export const PROVIDER_TYPE = Object.keys(PROVIDER_OBJ).map(key => {
  return {
    key,
    name: PROVIDER_OBJ[key],
    children: MODEL_TYPE.filter(item =>
      PROVIDER_MODEL_KEY[key]
        ? PROVIDER_MODEL_KEY[key].includes(item.key)
        : false,
    ),
  };
});

export const DEFAULT_CALLING = 'noSupport';
export const FUNC_CALLING = [
  { key: 'noSupport', name: i18n.t('modelAccess.noSupport') },
  { key: 'toolCall', name: 'Tool call' },
  /*{key: 'functionCall', name: 'Function call'},*/
];

export const DEFAULT_SUPPORT = 'noSupport';
export const SUPPORT_LIST = [
  { key: 'noSupport', name: i18n.t('modelAccess.noSupport') },
  { key: 'support', name: i18n.t('modelAccess.support') },
];

export const TYPE_OBJ = {
  apiKey: {
    [YUAN_JING]: 'sk-abc********************xyz',
    [OPENAI_API]: 'sk_7e4*************4s-BpI1l',
    [OLLAMA]: '',
    [QWEN]: 'sk-b************c70d',
    [HUOSHAN]: 'd8008ac0-****-****-****-**************',
    [INFINI]: 'sk-nw****gzjb6',
    [DEEPSEEK]: 'sk-14082***********************5e95',
    [QIANFAN]: 'bce-v3/ALTAK******82d1',
    [JINA]: 'jina_c08*********wMm',
  },
  inferUrl: {
    [`${ASR}_${QWEN}`]: 'https://dashscope.aliyuncs.com/api/v1',
    [`${ASR}_${HUOSHAN}`]:
      'https://openspeech.bytedance.com/api/v3/auc/bigmodel/recognize/flash',
    [`${MULTIMODAL_EMBEDDING}_${YUAN_JING}`]: i18n.t('modelAccess.noInferUrl'),
    [`${MULTIMODAL_RERANK}_${YUAN_JING}`]:
      'https://maas-api.ai-yuanjing.com/openapi/v1/yuanjing/reranker',
    [`${ASR}_${YUAN_JING}`]:
      'https://maas-api.ai-yuanjing.com/openapi/synchronous/asr/audio/file/transfer/unicom/sync/file/asr',
    [`${OCR}_${YUAN_JING}`]: 'https://maas-api.ai-yuanjing.com/openapi/v1',
    [`${GUI}_${YUAN_JING}`]: 'https://maas-api.ai-yuanjing.com/openapi/v1',
    [`${PDF_PARSER}_${YUAN_JING}`]:
      'https://maas-api.ai-yuanjing.com/openapi/v1',
    [YUAN_JING]: 'https://maas.ai-yuanjing.com/openapi/compatible-mode/v1',
    [OPENAI_API]: 'https://api.siliconflow.cn/v1',
    [OLLAMA]: 'https://192.168.21.100:11434/v1',
    [QWEN]: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    [HUOSHAN]: 'https://ark.cn-beijing.volces.com/api/v3',
    [INFINI]: 'https://cloud.infini-ai.com/maas/v1',
    [DEEPSEEK]: 'https://api.deepseek.com/v1',
    [QIANFAN]: 'https://qianfan.baidubce.com/v2',
    [JINA]: 'https://api.jina.ai/v1',
  },
};

export const IMAGE = 'image';
export const VIDEO = 'video';
export const SUPPORT_FILE_TYPE_OBJ = {
  [IMAGE]: i18n.t('modelAccess.supportFileType.pic'),
  [VIDEO]: i18n.t('modelAccess.supportFileType.video'),
};
