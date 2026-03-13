import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

// 获取模型统计数据
export const getModelData = params => {
  return service({
    url: `${USER_API}/statistic/model`,
    method: 'get',
    params,
  });
};

// 获取模型列表
export const fetchModelList = params => {
  return service({
    url: `${USER_API}/statistic/model/list`,
    method: 'get',
    params,
  });
};

// 模型数据导出
export const exportModelData = params => {
  return service({
    url: `${USER_API}/statistic/model/export`,
    method: 'get',
    params,
    responseType: 'blob',
  });
};
