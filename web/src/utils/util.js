import router from '@/router/index';
import { menuList } from '@/views/layout/menu';
import { checkPerm, PERMS } from '@/router/permission';
import { i18n } from '@/lang';
import { Message } from 'element-ui';
import { basePath } from '@/utils/config';

export function guid() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    let r = (Math.random() * 16) | 0,
      v = c == 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

export const getXClientId = () => localStorage.getItem('xClientId');

// 用于登录切组织等找到有权限的第一个菜单路径
export const fetchPermFirPath = (list = menuList) => {
  if (!list.length) return '';

  let path = '';
  for (let i in list) {
    const item = list[i];

    if (checkPerm(item.perm)) {
      if (item.children && item.children.length) {
        path = fetchPermFirPath(item.children).path;
        break;
      } else {
        path = item.path || '';
        break;
      }
    }
  }

  // 若有权限，跳转左侧菜单第一个有权限的页面；否则跳转 /404
  return { path: path || '/404' };
};

// 找到有权限的第一个菜单的 index
export const fetchCurrentPathIndex = (path, list) => {
  let index = '';
  const findIndex = list => {
    for (let i in list) {
      let item = list[i];
      const formatPath = url => {
        // 对于 文本问答/工作流/智能体 前面带了 /appSpace 特殊路由的处理
        if (url.includes('/appSpace/')) {
          return url.slice(9) + '/';
        }
        return url + '/';
      };
      if (item.path && formatPath(path).includes(formatPath(item.path))) {
        index = item.index;
      } else {
        if (item.children && item.children.length) {
          findIndex(item.children);
        }
      }
    }
    return index;
  };
  return findIndex(list);
};

export const jumpPermUrl = () => {
  const { path } = fetchPermFirPath();

  router.push({ path: path || '/404' });
};

export const jumpOAuth = params => {
  router.push({
    path: '/oauth',
    query: params,
  });
};

export const redirectUrl = () => {
  // 跳到有权限的第一个页面
  jumpPermUrl();
};

export const redirectUserInfoPage = (
  isUpdatePassword,
  callback,
  isRedirectUrl,
) => {
  if (isUpdatePassword !== undefined && !isUpdatePassword) {
    router.push('/userInfo?showPwd=1');
    callback && callback();
  } else {
    if (isRedirectUrl) jumpPermUrl();
  }
};

export const replaceIcon = logoPath => {
  let link =
    document.querySelector("link[rel*='icon']") ||
    document.createElement('link');
  link.type = 'image/x-icon';
  link.rel = 'shortcut icon';
  link.href = logoPath ? avatarSrc(logoPath) : basePath + '/aibase/favicon.ico';
  document.getElementsByTagName('head')[0].appendChild(link);
};

export const replaceTitle = title => {
  document.title = title || i18n.t('header.title');
};

export const copy = text => {
  let textareaEl = document.createElement('textarea');
  textareaEl.setAttribute('readonly', 'readonly'); // 防止手机上弹出软键盘
  textareaEl.value = text;
  document.body.appendChild(textareaEl);
  textareaEl.select();
  const res = document.execCommand('copy');
  document.body.removeChild(textareaEl);
  return res;
};

export const copyCb = () => {
  Message.success(i18n.t('common.copy.success'));
};

export const getInitTimeRange = () => {
  const date = new Date();
  const month = date.getMonth() + 1;
  const startTime =
    date.getFullYear() +
    '-' +
    (month < 10 ? '0' : '') +
    month +
    '-' +
    '01 00:00:00';
  const stamp = new Date().getTime() + 8 * 60 * 60 * 1000;
  const endTime = new Date(stamp)
    .toISOString()
    .replace(/T/, ' ')
    .replace(/\..+/, '')
    .substring(0, 19);
  return [startTime, endTime];
};

export function convertLatexSyntax(inputText) {
  // 1. 匹配块级公式，将 `\[` 和 `\]` 替换为 `$$`，支持 `\\[` `\\]` 或单个 `\[` `\]`
  inputText = inputText.replace(
    /\\\[\s*([\s\S]+?)\s*\\\]/g,
    (_, formula) => `$$${formula}$$`,
  );
  // 2. 匹配行内公式，将 `\(` 和 `\)` 替换为 `$`，支持 `\\(` `\\)` 或单个 `\(` `\)`
  inputText = inputText.replace(
    /\\\(\s*([\s\S]+?)\s*\\\)/g,
    (_, formula) => `$${formula}$`,
  );
  return inputText;
}

export function formatTimestamp(timestamp, format = 'YYYY-MM-DD HH:mm:ss') {
  const date = new Date(timestamp || timestamp);

  const map = {
    YYYY: date.getFullYear(),
    MM: String(date.getMonth() + 1).padStart(2, '0'),
    DD: String(date.getDate()).padStart(2, '0'),
    HH: String(date.getHours()).padStart(2, '0'),
    mm: String(date.getMinutes()).padStart(2, '0'),
    ss: String(date.getSeconds()).padStart(2, '0'),
  };

  return format.replace(/YYYY|MM|DD|HH|mm|ss/g, matched => map[matched]);
}

export function isSub(data) {
  return /\【([0-9]{0,2})\^\】/.test(data);
}

export function parseSub(data, index, searchList) {
  return data.replace(/\【([0-9]{0,2})\^\】/g, item => {
    let result = item.match(/\【([0-9]{0,2})\^\】/)[1];
    return `<sup class='citation' data-parents-index='${index}'>${result}</sup>`;
  });
  /*if (!searchList || !Array.isArray(searchList)) {
    searchList = [];
  }
  const result = data.match(/\【([0-9]{0,2})\^\】/g);
  if (!result) return data;
  return data.replace(/\【([0-9]{0,2})\^\】/g, item => {
    const num = item.replace(/\【|\^\】/g, '');
    if (!num) return item;
    const searchItem = searchList[Number(num)-1];
    if (!searchItem) return item;
    const snippet = searchItem ? searchItem.snippet : '';
    const title = searchItem ? searchItem.title : '';
    const displaySnippet = snippet.length >= 25 ? snippet.substring(0, 25) + '...' : snippet;
    return `
      <div class="citation-container" data-citation-index="${index}" data-citation-number="${num}">
        <sup class='citation' data-parents-index="${index}">${num}</sup>
        <div class="citation-tips">
          <div class="citation-tips-content">
            <div class="citation-tips-content-text">${displaySnippet}</div>
          </div>
          <div class="citation-tips-title">
            <span>
              <span class="el-icon-document"></span>
              <span>${title}</span>
            </span>
            <span class="el-icon-arrow-right citation-tips-content-icon" data-index="${index}" data-citation="${num}"></span>
          </div>
        </div>
      </div>
    `;
  });*/
}

// 子会话专用的 parseSub
export function parseSubConversation(text, index, searchList, id) {
  return text.replace(/\【([0-9]{0,2})\^\】/g, item => {
    let result = item.match(/\【([0-9]{0,2})\^\】/)[1];
    return `<sup class='citation' data-parents-index='${index}' data-pid='${id}'>${result}</sup>`;
  });
}

/**
 *获取URL参数
 */
export function getQueryString(val, href) {
  const hrefNew = href || window.location.href;
  const search = hrefNew.substring(
    hrefNew.lastIndexOf('?') + 1,
    hrefNew.length,
  );
  // 组装?
  const uri = '?' + search;
  const reg = new RegExp('' + val + '=([^&?]*)', 'ig');
  const matchArr = uri.match(reg);
  if (matchArr && matchArr.length) {
    return matchArr[0].substring(val.length + 1);
  }
  return null;
}

// 是否是有效的URL
export function isValidURL(string) {
  const res = string.match(
    /(https?|ftp|file|ssh):\/\/[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|]/i,
  );
  return res !== null;
}

export function isExternal(path) {
  return /^(https?:|mailto:|tel:)/.test(path);
}

export const formatTools = tools => {
  if (!(tools && tools.length)) return [];
  const newTools = tools.map((n, i) => {
    let params = [];
    let properties = n.inputSchema.properties;
    for (let key in properties) {
      params.push({
        name: key,
        requiredBadge:
          n.inputSchema.required && n.inputSchema.required.includes(key)
            ? i18n.t('common.required')
            : '',
        type: properties[key].type,
        description: properties[key].description,
      });
    }
    return {
      ...n,
      params,
    };
  });
  return newTools;
};

/**
 * 格式化得分，保留5位小数
 * @param {number|string} score - 得分值
 * @returns {string} 格式化后的得分字符串
 */
export function formatScore(score) {
  // 格式化得分，保留5位小数
  if (typeof score !== 'number') {
    return '0.00000';
  }
  return score.toFixed(5);
}

export function avatarSrc(path, defaultImg = '') {
  return path ? basePath + '/user/api/' + path : defaultImg;
}

// 换算单位万/亿/万亿，保留2位小数
export const formatAmount = (
  num,
  returnType = 'string',
  preserveRange = false,
) => {
  const units = i18n.t('statisticsEcharts.units');
  const isHasDecimal = num.toString().includes('.');
  let formatNum = num;
  let simplifiedNum = num.toString();

  // 99999以内原样显示
  if (preserveRange && num < 100000) {
    if (returnType === 'object') {
      return {
        value: simplifiedNum,
        type: '',
      };
    } else {
      return simplifiedNum;
    }
  }

  if (isHasDecimal) {
    formatNum = Number(num.toString().slice(0, num.toString().indexOf('.')));
  }
  // 获取数字的数量级
  let unitIndex = Math.floor((String(formatNum).length - 1) / 4);

  if (unitIndex > 0) {
    const unit = units[unitIndex];

    const divisor = Math.pow(10, unitIndex * 4);
    //缩小相应倍数，并保留2位小数
    const formattedValue = (num / divisor)
      .toFixed(2)
      .replace(/(\d)(?=(\d{3})+(?!\d))/g, '$1,');

    if (returnType === 'object') {
      return {
        value: formattedValue,
        type: unit,
      };
    } else {
      simplifiedNum = formattedValue + unit;
    }
  } else if (returnType === 'object') {
    // 数量级为0时的对象格式返回
    return {
      value: simplifiedNum,
      type: '',
    };
  }

  return simplifiedNum;
};

export function deepMerge(obj1, obj2) {
  for (let key in obj2) {
    if (obj2[key] && typeof obj2[key] === 'object') {
      if (!obj1[key] || typeof obj1[key] !== 'object') {
        obj1[key] = {};
      }
      deepMerge(obj1[key], obj2[key]);
    } else {
      obj1[key] = obj2[key];
    }
  }
  return obj1;
}

/**
 * 防抖函数（Debounce）
 * 限制函数在一定时间内的执行频率，合并短时间内的多次调用为一次
 * @param {Function} func - 需要防抖的函数
 * @param {number} wait - 等待时间（毫秒）
 * @param {boolean} immediate - 是否立即执行
 * @returns {Function} 防抖处理后的函数
 */
export function debounce(func, wait, immediate) {
  let timeout, args, context, timestamp, result;

  const later = function () {
    // 计算上次调用时间与当前时间的差值
    const last = +new Date() - timestamp;

    // 如果上次调用时间与当前时间的差值小于wait，则设置新的定时器
    if (last < wait && last >= 0) {
      timeout = setTimeout(later, wait - last);
    } else {
      // 否则执行函数
      timeout = null;
      if (!immediate) {
        result = func.apply(context, args);
        if (!timeout) context = args = null;
      }
    }
  };

  return function () {
    context = this;
    args = arguments;
    timestamp = +new Date();

    // 如果immediate为true且当前没有定时器，则立即执行函数
    const callNow = immediate && !timeout;

    // 设置定时器
    if (!timeout) {
      timeout = setTimeout(later, wait);
    }

    // 如果需要立即执行，则立即调用函数
    if (callNow) {
      result = func.apply(context, args);
      context = args = null;
    }

    return result;
  };
}

// 获取文件icon
export function getFileIcon(type) {
  switch (type) {
    case 'txt':
      return require('@/assets/imgs/txt-icon.png');
    case 'csv':
      return require('@/assets/imgs/csv-icon.png');
    case 'xlsx':
      return require('@/assets/imgs/xls-icon.png');
    case 'docx':
      return require('@/assets/imgs/word-icon.png');
    case 'pptx':
      return require('@/assets/imgs/ppt-icon.png');
    case 'pdf':
      return require('@/assets/imgs/pdf-icon.png');
    default:
      return require('@/assets/imgs/fileicon.png');
  }
}

// 文件大小格式化
export function formatFileSize(bytes, decimals = 2) {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return (
    parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[i]
  );
}

export function Md2Img(markdownText, escapeHtml = true) {
  // 匹配 Markdown 图片语法的正则表达式
  // ![](image.jpg) 或 ![alt](image.jpg) 或 ![alt](image.jpg "title")
  const imageRegex = /!\[(.*?)\]\(([^)\s]+)(?:\s+"([^"]*)")?\)/g;
  // 匹配 Markdown 换行符的正则表达式
  const newlineRegex = /(\r\n|\r|\n)/g;

  // 转义HTML特殊字符
  if (escapeHtml)
    markdownText = markdownText
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');

  let lastIndex = 0;
  let result = '';

  let match;
  while ((match = imageRegex.exec(markdownText)) !== null) {
    // 添加匹配前的文本内容
    result += markdownText.substring(lastIndex, match.index);

    // 构造图片HTML
    const alt = match[1] || '';
    const src = match[2];
    const title = match[3] ? ` title="${match[3]}"` : '';

    result += `<img src="${src}" alt="${alt}"${title}>`;

    // 更新lastIndex到匹配结束位置
    lastIndex = match.index + match[0].length;
  }

  // 添加剩余的文本内容
  result += markdownText.substring(lastIndex);

  // 将换行符转换为<br>标签
  result = result.replace(newlineRegex, '<br>');

  return result;
}

export function Img2Md(htmlString, escapeHtml = true) {
  if (['<div><br></div>', '<br>'].includes(htmlString)) return '';
  // 匹配 img 标签的正则表达式
  const imgRegex = /<img\s+[^>]*src\s*=\s*["']([^"']+)["'][^>]*>/gi;

  // 替换 img 标签为 Markdown 格式
  let result = htmlString.replace(imgRegex, (match, src) => {
    // 提取 alt 属性（如果有）
    const altMatch = match.match(/alt\s*=\s*["']([^"']*)["']/i);
    const alt = altMatch ? altMatch[1] : '';
    return `![${alt}](${src})`;
  });

  result = result
    // 处理空行
    .replace(/<div><br><\/div>/gi, '\n')
    // 处理块级元素的换行 - 仅在块级元素前添加换行符，后截替换为空
    .replace(/<(div|p|h[1-6]|li|blockquote)\b[^>]*>(.*?)<\/\1>/gi, '\n$2')
    // 处理自闭合的br标签
    .replace(/<br\s*\/?>/gi, '\n')
    // 删除所有其他HTML标签，只保留纯文本内容和换行符
    .replace(/<[^>]*>/g, '');

  // 恢复HTML特殊字符
  if (escapeHtml)
    result = result
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&quot;/g, '"')
      .replace(/&#39;/g, "'")
      .replace(/&amp;/g, '&');

  return result;
}
