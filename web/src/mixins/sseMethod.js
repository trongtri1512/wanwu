import { fetchEventSource } from '../sse/index.js';
import { store } from '@/store/index';
import Print from '../utils/printPlus2.js';
import {
  parseSub,
  convertLatexSyntax,
  parseSubConversation,
} from '@/utils/util.js';
import { mapActions, mapGetters } from 'vuex';
import { i18n } from '@/lang';
import StreamProcessor from '@/utils/streamProcessor.js';

var originalFetch = window.fetch;

import { md } from './markdown-it';
import $ from './jquery.min.js';
import { OPENURL_API, USER_API } from '@/utils/requestConstants';
import { getCustomSkillSSeUrl } from '@/api/templateSquare';

const AGENT_API_URL = `${USER_API}/assistant/stream`;
const RAG_API_URL = `${USER_API}/rag/chat`;
const EXPRIENCE_API_URL = `${USER_API}/model/experience/llm`;

export default {
  data() {
    return {
      isTestChat: false,
      defaultUrl: '/img/smart/logo.png',
      inputVal: '',
      eventSource: null,
      ctrlAbort: null,
      sseParams: {},
      sseResponse: {},
      echo: true,
      conversationId: '', //会话id
      chatList: [],
      reminderList: [],
      queryFilePath: '',
      stopBtShow: false,
      origin: window.location.origin,
      reconnectCount: 0,
      isEnd: true,
      sseApi: AGENT_API_URL,
      rag_sseApi: RAG_API_URL,
      exprience_sseApi: EXPRIENCE_API_URL,
      lastIndex: 0,
      query: '',
      isStoped: false,
      access_token: '',
      runResponse: '',
      fileList: [], // 文件列表
      instanceSessionStatus: -1,
      sessionComRef: null,
      _subConversionsMap: null, // 子会话存储 Map
      _subConversionProcessors: null, // 子会话处理器 Map
      responseFiles: [], // 用于存储 SSE 返回的附件文件列表
    };
  },
  created() {
    if (!this.isExplorePage()) {
      this.rag_sseApi = `${RAG_API_URL}/draft`;
    }
    const vuex = JSON.parse(localStorage.getItem('access_cert'));
    if (vuex) {
      this.access_token = vuex.user.token;
    }
  },
  mounted() {
    //this.addVisibilitychangeEvent()
  },
  beforeDestroy() {
    this.setStoreSessionStatus(-1);
    this.stopEventSource();
    this._print && this._print.stop();
  },
  computed: {
    ...mapGetters('app', ['sessionStatus']),
    ...mapGetters('user', ['token', 'userInfo']),
  },
  methods: {
    ...mapActions('app', ['setStoreSessionStatus']),
    isExplorePage() {
      return this.$route.path.includes('/explore/');
    },
    newFetch(url, options) {
      // 可以调用原始的 fetch 函数
      if (this.isStoped) {
        return;
      }
      return originalFetch(url, options)
        .then(response => {
          // 可以在这里修改响应或者添加额外的处理
          let query = this.query;

          if (response.status != 200) {
            let me = this;
            try {
              const stream = response.body;

              const reader = stream.getReader();
              const decoder = new TextDecoder('utf-8');

              function readStream() {
                reader
                  .read()
                  .then(({ done, value }) => {
                    if (done) {
                      console.log('Stream complete');
                      reader.releaseLock();
                      return;
                    }

                    // Decode and process each chunk of data.
                    const decodedValue = decoder.decode(value, {
                      stream: true,
                    });

                    if (decodedValue) {
                      let msg = JSON.parse(decodedValue).msg;
                      me.setStoreSessionStatus(-1);
                      var fillData = {
                        query: query,
                        qa_type: 0,
                        finish: 1,
                        response: msg, //非代码文本使用自定义转换规则，不使用markdown,(markdown渲染会导致卡顿或样式丢失)
                        oriResponse: '',
                        searchList: [], //过滤包含yunyingshang文件的出处
                      };
                      this.runResponse = msg;
                    }
                    readStream();

                    // Continue reading the stream.
                  })
                  .catch(err => {
                    console.error('Reading stream failed1:', err);
                  });
              }

              readStream();
              me.isStoped = true;
            } catch (e) {
              console.error('Reading stream failed:', e);
            }
          }

          return response;
        })
        .catch(err => {
          this.$message.warning(i18n.t('sse.connectError'));
          this.isEnd = true;
          this.setStoreSessionStatus(-1);
          this.runDisabled = false;
        });
    },
    ...mapActions('app', ['setStoreSessionStatus']),
    queryCopy(text) {
      this.setPrompt(text);
    },
    /*过滤掉markdown中自定义的行号*/
    getContentInBraces(shtml) {
      let temp = document.createElement('div');
      temp.setAttribute('id', 'temp');
      temp.innerHTML = shtml;
      document.body.appendChild(temp);
      $(temp).find('.line-num').remove();
      return temp.innerText;
    },
    // 填充开场白
    setProloguePrompt(val) {
      // this.$refs['editable'].setPrompt(val)
      const editable =
        this.$refs.editable || (this.getEditableRef && this.getEditableRef());
      if (editable) {
        editable.setPrompt(val);
      }
      this.preSend();
    },
    //获取上传的文件
    getFileIdList() {
      const editable =
        this.$refs.editable || (this.getEditableRef && this.getEditableRef());
      let list = editable.getFileIdList();
      let fileIds = [];
      this.queryFilePath = '';
      if (list.length) {
        fileIds = list.map(n => {
          return n.fileId;
        });
        this.queryFilePath = list[0].url;
      }
      return fileIds.join(',');
    },
    mouseEnter(n) {
      n.hover = true;
    },
    mouseLeave(n) {
      n.hover = false;
    },
    setSessionStatus(status) {
      // this.setStoreSessionStatus(status)
      if (this.fieldId) {
        this.instanceSessionStatus = status;
      } else {
        this.setStoreSessionStatus(status);
      }
    },
    getCurrentSessionStatus() {
      return this.fieldId ? this.instanceSessionStatus : this.sessionStatus;
    },
    setSseParams(data) {
      // this.sseParams = data

      this.sseParams = data ? Object.assign({}, data) : {};
      if (data && data.sessionComRef) {
        this.sessionComRef = data.sessionComRef;
      }
    },
    fetchEventSource(url, params, options = {}) {
      const {
        onopen,
        onmessage,
        onclose = () => {
          console.log('===> eventSource onClose');
          this.setStoreSessionStatus(-1); //关闭后改变状态
          this.sseOnCloseCallBack();
        },
        onerror = e => {
          console.log(i18n.t('sse.connectError'));
          if (e.readyState === EventSource.CLOSED) {
            console.log('connection is closed');
          } else {
            console.warn('Error occured', e);
          }
          this.stopEventSource(); //前端主动关闭连接
          this.setStoreSessionStatus(-1); //关闭后改变状态
        },
        headers,
        signal,
        ...rest
      } = options;
      this.ctrlAbort = new AbortController();
      return new fetchEventSource(this.origin + url, {
        method: 'POST',
        headers: headers || {
          'Content-Type': 'application/json',
          Authorization: 'Bearer ' + this.token,
          'x-user-id': this.userInfo.uid,
          'x-org-id': this.userInfo.orgId,
        },
        signal: signal || this.ctrlAbort.signal,
        body: JSON.stringify(params),
        openWhenHidden: true,
        onopen: onopen,
        onmessage: onmessage,
        onclose: onclose,
        onerror: onerror,
        rest,
      });
    },
    doragSend() {
      this.stopBtShow = true;
      this.isStoped = false;
      let _history = this.$refs['session-com'].getList();
      this.sendEventStream(this.inputVal, '', _history.length);
    },
    sendEventStream(prompt, msgStr, lastIndex) {
      let sessionCom = this.sessionComRef || this.$refs['session-com'];
      if (!sessionCom) {
        console.warn('[sseMethod] session-com ref missing');
        return;
      }
      if (this.getCurrentSessionStatus() === 0) {
        this.$message.warning(i18n.t('sse.incompleteError'));
        return;
      }

      this.sseResponse = {};
      this.setStoreSessionStatus(0);
      this.clearInput();

      let params = {
        query: prompt,
        pending: true,
        responseLoading: true,
        requestFileUrls: [],
        fileList: this.fileList,
        pendingResponse: '',
      };
      sessionCom.pushHistory(params);

      // 初始化流处理器
      const processor = new StreamProcessor({
        lastIndex,
        md,
        parseSub,
        convertLatexSyntax,
      });

      this._print = new Print({
        onPrintEnd: () => {
          this.onMainPrintEnd && this.onMainPrintEnd();
        },
      });
      let history_list = sessionCom.getSessionData();
      const history =
        history_list['history'].length > 1
          ? history_list['history'][history_list['history'].length - 2][
              'history'
            ]
          : [];

      this.eventSource = this.fetchEventSource(
        this.rag_sseApi,
        { ...this.sseParams, history: history },
        {
          onopen: async e => {
            if (e.status !== 200) {
              try {
                const errorData = await e.json();
                let commonData = {
                  ...this.sseParams,
                  query: prompt,
                };
                let fillData = {
                  ...commonData,
                  response: errorData.msg,
                };
                sessionCom.replaceLastData(lastIndex, fillData);
              } catch (e) {
                const text = await e.text();
                this.$message.error(text || i18n.t('sse.error'));
              }

              this.stopEventSource();
              this.setStoreSessionStatus(-1);
            }
          },
          onmessage: e => {
            if (e && e.data) {
              let data;
              try {
                data = JSON.parse(e.data);
              } catch (error) {
                return; // 如果解析失败，直接返回，不处理这条消息
              }

              this.sseResponse = data;
              let commonData = {
                ...this.sseResponse,
                ...this.sseParams,
                query: prompt,
                fileList: this.fileList,
                response: '',
                filepath: data.file_url || '',
                requestFileUrls: '',
                gen_file_url_list: [],
                searchList:
                  data.data && data.data.searchList ? data.data.searchList : [],
                thinkText: i18n.t('sse.thinkingText'),
                isOpen: true,
                citations: [],
              };

              if (data.code === 0 || data.code === 1) {
                //finish 0：进行中  1：关闭   2:敏感词关闭
                let _sentence = data.data.output;

                this._print.print(
                  {
                    response: _sentence,
                    finish: data.finish,
                  },
                  commonData,
                  (worldObj, search_list) => {
                    this.setStoreSessionStatus(0);
                    processor.updateSearchList(search_list);
                    processor.append(worldObj.world);

                    const renderResult = processor.getRenderResult();

                    let fillData = {
                      ...commonData,
                      ...renderResult,
                      finish: worldObj.finish,
                      searchList: search_list
                        ? search_list.map(n => ({
                            ...n,
                            snippet: n.snippet ? md.render(n.snippet) : '',
                          }))
                        : [],
                    };

                    if (worldObj.finish === 2) {
                      fillData.response = this.$t('sse.sensitiveTips');
                      sessionCom.replaceLastData(lastIndex, fillData);
                      this.$nextTick(() => sessionCom.scrollBottom());
                      this.setStoreSessionStatus(-1);
                    } else {
                      sessionCom.replaceLastData(lastIndex, fillData);
                    }

                    if (worldObj.isEnd && worldObj.finish === 1) {
                      this.setStoreSessionStatus(-1);
                    }
                  },
                );
              } else if (data.code === 7 || data.code === -1) {
                this.setStoreSessionStatus(-1);
                sessionCom.replaceLastData(lastIndex, {
                  ...commonData,
                  response: data.message,
                });
              }
            }
          },
        },
      );
    },
    doSend(params) {
      this.stopBtShow = true;
      this.isStoped = false;
      let _history = this.$refs['session-com'].getList();
      this.sendEventSource(this.inputVal, '', _history.length);
    },
    sendEventSource(prompt, msgStr, lastIndex) {
      console.log('####  sendEventSource', new Date().getTime());
      let sessionCom = this.sessionComRef || this.$refs['session-com'];
      if (!sessionCom) {
        console.warn('[sseMethod] session-com ref missing');
        return;
      }
      if (this.getCurrentSessionStatus() === 0) {
        this.$message.warning(i18n.t('sse.incompleteError'));
        return;
      }

      this.sseResponse = {};
      this.setStoreSessionStatus(0);
      this.clearInput();

      let params = {
        query: prompt,
        pending: true,
        responseLoading: true,
        requestFileUrls: this.queryFilePath ? [this.queryFilePath] : [],
        fileList: this.fileList,
        pendingResponse: '',
      };
      sessionCom.pushHistory(params);

      this._print = new Print({
        onPrintEnd: () => {
          this.onMainPrintEnd && this.onMainPrintEnd();
        },
      });

      let data = null;
      let headers = null;
      //判断是是不是openurl对话
      if (this.type === 'agentChat') {
        if (!this.isExplorePage()) {
          this.sseApi = `${AGENT_API_URL}/draft`;
        } else {
          this.sseApi = AGENT_API_URL;
        }
        data = {
          ...this.sseParams,
          prompt,
          systemPrompt: this.sseParams.systemPrompt, //提示词对比参数
        };
        headers = {
          'Content-Type': 'application/json',
          Authorization: 'Bearer ' + this.token,
          'x-user-id': this.userInfo.uid,
          'x-org-id': this.userInfo.orgId,
        };
      } else {
        this.sseApi = `${OPENURL_API}/agent/${this.sseParams.assistantId}/stream`;
        data = {
          conversationId: this.sseParams.conversationId,
          prompt,
        };
        headers = {
          'X-Client-ID': this.getHeaderConfig().headers['X-Client-ID'],
        };
      }

      this._subConversionsMap = new Map(); // 子会话数据Map
      this._subConversionProcessors = new Map(); // 每个 order 的子处理器
      this._mainProcessors = new Map(); // 每个 order 的主处理器

      this.eventSource = this.fetchEventSource(this.sseApi, data, {
        headers,
        ...(this.type === 'webChat' && { isOpenUrl: true }),
        onopen: async e => {
          console.log('已建立SSE连接~', new Date().getTime());
          if (e.status !== 200) {
            try {
              const errorData = await e.json();
              let commonData = {
                ...this.sseParams,
                query: prompt,
              };
              let fillData = {
                ...commonData,
                response: errorData.msg,
              };
              sessionCom.replaceLastData(lastIndex, fillData);
            } catch (e) {
              const text = await e.text();
              this.$message.error(text || i18n.t('sse.error'));
            }

            this.stopEventSource();
            this.setStoreSessionStatus(-1);
            return;
          }
        },
        onmessage: e => {
          if (e && e.data) {
            let data = JSON.parse(e.data);
            console.log('===>', new Date().getTime(), data);
            this.sseResponse = data;
            //待替换的数据，需要前端组装
            let commonData = {
              ...data,
              ...this.sseParams,
              query: prompt,
              fileList: this.fileList,
              response: '',
              filepath: data.file_url || '',
              requestFileUrls: this.queryFilePath
                ? [this.queryFilePath]
                : data.requestFileUrls,
              searchList: data.search_list || [],
              gen_file_url_list: data.gen_file_url_list || [],
              thinkText: i18n.t('agent.thinking'),
              toolText: '使用工具中...',
              isOpen: true,
              showScrollBtn: null,
              citations: [],
              subConversions: [], // 初始化子会话列表
              messageSequence: [], // 初始化消息序列，用于平铺渲染
              _lastOrder: -1, // 内部追踪最后一次的 order
            };

            if (data.code === 0) {
              // 处理子会话消息 (eventType === 0)
              if (data.eventType === 1 && data.eventData) {
                const { id, name, status, timeCost, profile } = data.eventData;
                let subConversion = this._subConversionsMap.get(id);
                let subProcessor = this._subConversionProcessors.get(id);

                if (!subConversion) {
                  subConversion = {
                    id,
                    name,
                    status, // 1开始、2输出中、3结束、4处理失败
                    timeCost,
                    profile, //头像
                    response: '',
                    stableChunks: [],
                    activeResponse: '',
                    isOpen: false, // 默认收起
                    searchList: data.search_list || [], // 初始化 searchList
                    citationsTagList: [], // 已引用的出处索引
                  };
                  this._subConversionsMap.set(id, subConversion);

                  // 初始化流处理器
                  subProcessor = new StreamProcessor({
                    lastIndex,
                    md,
                    parseSub: (text, index, searchList) =>
                      parseSubConversation(text, index, searchList, id),
                    convertLatexSyntax,
                    searchList: subConversion.searchList,
                  });
                  this._subConversionProcessors.set(id, subProcessor);
                } else {
                  // 更新状态和耗时
                  subConversion.status = status;
                  if (timeCost) subConversion.timeCost = timeCost;
                  // 如果后续包中有 search_list，则更新
                  if (data.search_list && data.search_list.length) {
                    subConversion.searchList = data.search_list;
                    subProcessor.updateSearchList(data.search_list);
                  }
                }

                // 累加回复内容并处理流
                if (data.response) {
                  // 处理转义换行符
                  let processedResponse = data.response.replace(/\\n/g, '\n');
                  subConversion.response += processedResponse;
                  subProcessor.append(processedResponse);
                  const renderResult = subProcessor.getRenderResult();
                  subConversion.stableChunks = renderResult.stableChunks;
                  subConversion.activeResponse = renderResult.activeResponse;
                  // StreamProcessor 增量维护的引文列表
                  subConversion.citationsTagList = renderResult.citations || [];
                }

                // 更新消息序列
                let sequence =
                  sessionCom.getSessionData()['history'][lastIndex]
                    .messageSequence || [];
                if (data.order !== undefined && data.order !== null) {
                  let currentSubItem = sequence.find(
                    item =>
                      item.type === 'sub' &&
                      item.id === id &&
                      item.order === data.order,
                  );
                  if (!currentSubItem) {
                    currentSubItem = {
                      type: 'sub',
                      id: id,
                      order: data.order,
                    };
                    sequence.push(currentSubItem);
                  }
                }

                // 构造 fillData
                // 获取最新的子会话列表
                const subConversionsList = Array.from(
                  this._subConversionsMap.values(),
                );

                let fillData = {
                  ...commonData,
                  finish:
                    this._currentMainFinish !== undefined
                      ? this._currentMainFinish
                      : 0,
                  subConversions: subConversionsList,
                  messageSequence: sequence,
                };
                sessionCom.replaceLastData(lastIndex, fillData);
                // 如果子智能体结束或失败，可能需要滚动到底部
                if (status === 3 || status === 4) {
                  this.$nextTick(() => sessionCom.scrollBottom());
                }
              } else {
                // 主智能体消息 (eventType === 0 或 undefined)
                // 更新当前主智能体 finish 状态
                this._currentMainFinish = data.finish;

                // 根据 order 获取或创建对应的 processor
                const currentOrder = data.order !== undefined ? data.order : 0;
                let mainProcessor = this._mainProcessors.get(currentOrder);

                if (!mainProcessor) {
                  mainProcessor = new StreamProcessor({
                    lastIndex,
                    md,
                    parseSub,
                    convertLatexSyntax,
                  });
                  this._mainProcessors.set(currentOrder, mainProcessor);
                }

                //finish 0：进行中  1：关闭   2:敏感词关闭
                let _sentence = data.response;
                this._print.print(
                  {
                    response: _sentence,
                    finish: data.finish,
                  },
                  commonData,
                  (worldObj, search_list) => {
                    this.setStoreSessionStatus(0);
                    mainProcessor.updateSearchList(search_list);
                    mainProcessor.append(worldObj.world);

                    const renderResult = mainProcessor.getRenderResult();

                    // 更新消息序列
                    let sequence =
                      sessionCom.getSessionData()['history'][lastIndex]
                        .messageSequence || [];

                    if (data.order !== undefined && data.order !== null) {
                      let currentMainItem = sequence.find(
                        item =>
                          item.type === 'main' && item.order === data.order,
                      );

                      if (!currentMainItem) {
                        currentMainItem = {
                          type: 'main',
                          order: data.order,
                          renderedContent: '',
                          stableChunks: [],
                          activeResponse: '',
                        };
                        sequence.push(currentMainItem);
                      }

                      currentMainItem.renderedContent = renderResult.response;
                      currentMainItem.stableChunks = renderResult.stableChunks;
                      currentMainItem.activeResponse =
                        renderResult.activeResponse;
                    }

                    // 获取最新的子会话列表
                    const subConversionsList = Array.from(
                      this._subConversionsMap.values(),
                    );

                    let fillData = {
                      ...commonData,
                      ...renderResult,
                      finish: worldObj.finish,
                      searchList:
                        search_list && search_list.length
                          ? search_list.map(n => ({
                              ...n,
                              snippet: md.render(n.snippet),
                            }))
                          : [],
                      subConversions: subConversionsList,
                      messageSequence: sequence,
                    };
                    sessionCom.replaceLastData(lastIndex, fillData);
                    if (worldObj.finish !== 0) {
                      if (worldObj.finish === 4) {
                        let fillData = {
                          ...commonData,
                          response: i18n.t('sse.sensitiveTips'),
                          subConversions: subConversionsList,
                          messageSequence: sequence,
                        };
                        sessionCom.replaceLastData(lastIndex, fillData);
                        this.$nextTick(() => {
                          sessionCom.scrollBottom();
                        });
                      }
                      this.setStoreSessionStatus(-1);
                    }

                    if (worldObj.isEnd && worldObj.finish === 1) {
                      this.setStoreSessionStatus(-1);
                      this._currentMainFinish = undefined;
                    }
                  },
                );
              }
            } else if (data.code === 7 || data.code === -1 || data.code === 1) {
              this.setStoreSessionStatus(-1);
              // 获取最新的子会话列表，防止被覆盖
              const subConversionsList = this._subConversionsMap
                ? Array.from(this._subConversionsMap.values())
                : [];
              let fillData = {
                ...commonData,
                response: data.message,
                subConversions: subConversionsList,
              };
              sessionCom.replaceLastData(lastIndex, fillData);
              this._currentMainFinish = undefined;
            }
          }
        },
      });
    },
    doExprienceSend(params) {
      this.stopBtShow = true;
      this.isStoped = false;
      let _history = this.$refs['session-com'].getList();
      this.sendExprienceEventStream(params.inputVal, '', _history.length);
    },
    sendExprienceEventStream(prompt, msgStr, lastIndex) {
      this.sseResponse = {};
      this.setStoreSessionStatus(0);
      let params = {
        query: prompt,
        pending: true,
        responseLoading: true,
        requestFileUrls: [],
        fileList: this.fileList,
        pendingResponse: '',
      };
      this.$refs['session-com'].pushHistory(params);
      let endStr = '';
      this._print = new Print({
        onPrintEnd: () => {
          // this.setStoreSessionStatus(-1)
        },
      });

      this.eventSource = this.fetchEventSource(
        this.exprience_sseApi,
        {
          ...this.apiParams,
          content: prompt,
        },
        {
          onopen: async e => {
            //console.log("已建立SSE连接~",new Date().getTime());
            if (e.status !== 200) {
              try {
                const errorData = await e.json();
                let commonData = {
                  ...this.sseParams,
                  query: prompt,
                };
                let fillData = {
                  ...commonData,
                  response: errorData.msg,
                };
                this.$refs['session-com'].replaceLastData(lastIndex, fillData);
              } catch (e) {
                const text = await e.text();
                this.$message.error(text || i18n.t('sse.error'));
              }

              this.stopEventSource();
              this.setStoreSessionStatus(-1);
              return;
            }
          },
          onmessage: e => {
            if (e && e.data) {
              let data;
              try {
                data = JSON.parse(e.data);
                // console.log('===>', new Date().getTime(), data);
              } catch (error) {
                return; // 如果解析失败，直接返回，不处理这条消息
              }
              if (
                Array.isArray(data.choices) &&
                data.choices[0] &&
                data.choices[0].delta
              ) {
                data.response = data.choices[0].delta.content;
                data.finish =
                  data.choices[0].finish_reason === 'stop' ||
                  data.choices[0].delta.content === 'stop';
              } else {
                data.response = '';
                data.finish = true;
              }
              this.setStoreSessionStatus(0);
              this.sseResponse = data;
              //待替换的数据，需要前端组装
              let commonData = {
                ...this.sseResponse,
                ...this.sseParams,
                query: prompt,
                fileName: '',
                fileSize: '',
                response: '',
                filepath: '',
                requestFileUrls: '',
                searchList:
                  this.sseResponse.data && this.sseResponse.data.searchList
                    ? this.sseResponse.data.searchList
                    : [],
                gen_file_url_list: [],
                thinkText: i18n.t('sse.thinkingText'),
                isOpen: true,
                citations: [],
                qa_type: 0, // 为了组件复用，前端加了标识
              };
              if ([7, -1].includes(data.code)) {
                this.setStoreSessionStatus(-1);
                let fillData = {
                  ...commonData,
                  response: data.message,
                };
                this.$refs['session-com'].replaceLastData(lastIndex, fillData);
              } else {
                //finish 0：进行中  1：关闭   2:敏感词关闭
                this._print.print(
                  {
                    response: data.response,
                    finish: data.finish,
                  },
                  commonData,
                  (worldObj, search_list) => {
                    this.setStoreSessionStatus(0);
                    endStr += worldObj.world;
                    endStr = convertLatexSyntax(endStr);
                    endStr = parseSub(endStr, lastIndex);
                    let fillData = {
                      ...commonData,
                      response: md.render(endStr),
                      oriResponse: endStr,
                      finish: worldObj.finish,
                      searchList:
                        search_list && search_list.length
                          ? search_list.map(n => ({
                              ...n,
                              snippet: n.snippet ? md.render(n.snippet) : '',
                            }))
                          : [],
                    };
                    this.$refs['session-com'].replaceLastData(
                      lastIndex,
                      fillData,
                    );
                    if (worldObj.isEnd && worldObj.finish) {
                      this.setStoreSessionStatus(-1);
                    }
                  },
                );
              }
            }
          },
        },
      );
    },
    // 多线程SSE简化版本
    sendEventStreamIsolation(url, params, callbacks = {}, timeout = 0) {
      let fullContent = '';
      let isCompleted = false;
      const { onProgress, onComplete } = callbacks;

      const _print = new Print({});
      const ctrlAbort = new AbortController();

      const handleComplete = content => {
        if (isCompleted) return;
        isCompleted = true;
        ctrlAbort.abort();
        if (onComplete) onComplete(content);
      };

      this.fetchEventSource(`${USER_API}` + url, params, {
        onopen: async response => {
          if (response.status !== 200) {
            try {
              const errorData = await response.json();
              console.log('Network error', errorData);
              this.$message.error(errorData.msg || i18n.t('sse.error'));
            } catch (e) {
              console.error('Failed to parse error response', e);
              this.$message.error(i18n.t('sse.error'));
            }
            handleComplete(fullContent);
          }
        },
        onmessage: e => {
          if (e && e.data) {
            try {
              const data = JSON.parse(e.data);
              _print.print(
                {
                  response: data.response,
                  finish: data.finish,
                },
                {},
                worldObj => {
                  fullContent += worldObj.world;
                  if (onProgress) onProgress(fullContent, worldObj);
                  if (Boolean(worldObj.finish)) {
                    console.log('===> eventSource onComplete');
                    handleComplete(fullContent);
                  }
                },
              );
            } catch (e) {
              console.warn('message json parse fail: ', e);
            }
          }
        },
        onclose: () => {
          console.log('===> eventSource onClose');
          handleComplete(fullContent);
        },
        onerror: e => {
          console.log(i18n.t('sse.connectError'));
          if (e.readyState === EventSource.CLOSED) {
            console.log('connection is closed');
          } else {
            console.warn('Error occured', e);
          }
          handleComplete(fullContent);
        },
        signal: ctrlAbort.signal,
      });

      if (timeout > 0) {
        setTimeout(() => {
          if (!ctrlAbort.signal.aborted) {
            ctrlAbort.abort();
            this.$message.warning(i18n.t('sse.timeoutError'));
            handleComplete(fullContent);
          }
        }, timeout);
      }
    },
    preStop() {
      //获取已经拿到的全部回答,一次性回显出来
      this.sseOnCloseCallBack(true);
    },
    sseOnCloseCallBack(isStoped) {
      this.stopEventSource();
      //图文问答不使用打字机
      /* if(this.sseResponse.qa_type === 6){
                return
            }*/
      //主动停止
      if (isStoped) {
        // 手动停止时，将所有进行中的子会话状态置为失败/停止
        if (this._subConversionsMap) {
          let hasUpdate = false;
          for (let sub of this._subConversionsMap.values()) {
            if (sub.status === 1 || sub.status === 2) {
              sub.status = 4;
              hasUpdate = true;
            }
          }
          if (hasUpdate) {
            let sessionCom = this.sessionComRef || this.$refs['session-com'];
            if (sessionCom) {
              let history = sessionCom.getSessionData().history;
              let lastIndex = history.length - 1;
              if (lastIndex >= 0) {
                const subConversionsList = Array.from(
                  this._subConversionsMap.values(),
                );
                let lastItem = history[lastIndex];
                sessionCom.replaceLastData(lastIndex, {
                  ...lastItem,
                  subConversions: subConversionsList,
                });
              }
            }
          }
        }
        this.stopAndEcho();
      } else {
        //收到onclose,且使用的是文生代码
        if (this.sseResponse.qa_type === 4) {
          this.stopAndEcho();
        } else {
          //接口405等
          let history_list = [];
          let lastIndex = history_list.length - 1;
          let lastRQ = history_list[lastIndex];
          let endStr = this._print.getAllworld();
          endStr = convertLatexSyntax(endStr);
          // 替换标签
          endStr = parseSub(endStr);
          // 如果返回有结果，则在结束时不展示“本次回答已终止”
          this.runResponse = md.render(endStr);
          this.runDisabled = false;
          this.setStoreSessionStatus(-1);
        }
      }
    },
    stopAndEcho() {
      //暂存已经收到的所有response
      let endResponse = this._print.getAllworld();

      this._print && this._print.stop();

      setTimeout(() => {
        this.setStoreSessionStatus(-1);

        let history_list = [];
        let lastIndex = history_list.length - 1;
        let lastRQ = history_list[lastIndex];
        if (endResponse) {
          endResponse = convertLatexSyntax(endResponse);
          // 替换标签
          endResponse = parseSub(endResponse);
          this.runResponse = md.render(endResponse);
          this.runDisabled = false;
        } else {
          if (
            Object.keys(this.sseResponse).length !== 0 &&
            this.sseResponse.code !== 7
          ) {
            this.runResponse = '本次回答已被终止';
            this.setStoreSessionStatus(-1);
          } else {
            this.stopEventSource();
            this.setStoreSessionStatus(-1);
            this.$refs['session-com'].stopPending();
          }
        }
      }, 15);
    },
    stopEventSource() {
      this.ctrlAbort && this.ctrlAbort.abort();
      this.eventSource = null;
    },
    refreshLastSession() {
      let endResponse = this._print.getAllworld();
      let history_list = [];
      let lastIndex = history_list.length - 1;
      let lastRQ = history_list[lastIndex];
      // this.$refs['session-com'].replaceLastData(lastIndex, {
      //     ...lastRQ,
      //     response: endResponse
      // })
    },
    setPrompt(data) {
      const editable =
        this.$refs.editable || (this.getEditableRef && this.getEditableRef());
      if (editable) {
        editable.setPrompt(data);
      }
      // this.$refs['editable'].setPrompt(data)
    },
    clearInput() {
      const editable =
        this.$refs.editable || (this.getEditableRef && this.getEditableRef());
      if (editable) {
        editable.clearInput();
        editable.clearFile();
      }
      this.inputVal = '';
      this.fileId = '';
    },
    clearPageHistory() {
      this.$refs['session-com'] && this.$refs['session-com'].clearData();
      // this.$refs.editable && this.clearInput()
      this.clearInput();
    },
    clearHistory() {
      this.stopBtShow = false;
      this.clearPageHistory();
    },
    refresh() {
      let sessionCom = this.sessionComRef || this.$refs['session-com'];
      if (!sessionCom) return;
      let history_list = sessionCom.getList();
      let _history = history_list[history_list.length - 1];
      let inputVal = _history.query;
      let fileInfo = _history.fileInfo ? _history.fileInfo : [];
      let fileList = _history.fileList ? _history.fileList : [];
      this.preSend(inputVal, fileList, fileInfo);
    },
    // skills创建会话发送
    doSkillsSend() {
      this.stopBtShow = true;
      this.isStoped = false;
      let _history = this.$refs['session-com'].getList();
      this.sendSkillEventSource(this.inputVal, '', _history.length);
    },
    // skills创建会话sse
    sendSkillEventSource(prompt, msgStr, lastIndex) {
      console.log('####  sendEventSource', new Date().getTime());
      let sessionCom = this.sessionComRef || this.$refs['session-com'];
      if (!sessionCom) {
        console.warn('[sseMethod] session-com ref missing');
        return;
      }
      if (this.getCurrentSessionStatus() === 0) {
        this.$message.warning(i18n.t('sse.incompleteError'));
        return;
      }

      this.sseResponse = {};
      this.responseFiles = []; // 重置附件列表
      this.setStoreSessionStatus(0);
      this.clearInput();

      let params = {
        query: prompt,
        pending: true,
        responseLoading: true,
        requestFileUrls: this.queryFilePath ? [this.queryFilePath] : [],
        fileList: this.fileList,
        pendingResponse: '',
      };
      sessionCom.pushHistory(params);

      this._print = new Print({
        onPrintEnd: () => {
          this.onMainPrintEnd && this.onMainPrintEnd();
        },
      });

      let data = null;
      let headers = null;
      //判断是是不是openurl对话
      if (this.type === 'agentChat') {
        this.sseApi = getCustomSkillSSeUrl();
        data = {
          ...this.sseParams,
          query: prompt,
        };
        headers = {
          'Content-Type': 'application/json',
          Authorization: 'Bearer ' + this.token,
          'x-user-id': this.userInfo.uid,
          'x-org-id': this.userInfo.orgId,
        };
      }

      this._subConversionsMap = new Map(); // 子会话数据Map
      this._subConversionProcessors = new Map(); // 每个 order 的子处理器
      this._mainProcessors = new Map(); // 每个 order 的主处理器

      function transformSkillData(rawData) {
        const { metadata, ...rest } = rawData;
        const result = { ...metadata };
        Object.keys(rest).forEach(key => {
          // 若metadata已经存在同名key，则外层key 加_前缀以区分
          if (key in metadata) {
            result[`_${key}`] = rest[key];
          } else {
            result[key] = rest[key];
          }
        });
        return result;
      }

      this.eventSource = this.fetchEventSource(this.sseApi, data, {
        headers,
        ...(this.type === 'webChat' && { isOpenUrl: true }),
        onopen: async e => {
          console.log('已建立SSE连接~', new Date().getTime());
          if (e.status !== 200) {
            try {
              const errorData = await e.json();
              let commonData = {
                ...this.sseParams,
                query: prompt,
              };
              let fillData = {
                ...commonData,
                response: errorData.msg,
              };
              sessionCom.replaceLastData(lastIndex, fillData);
            } catch (e) {
              const text = await e.text();
              this.$message.error(text || i18n.t('sse.error'));
            }

            this.stopEventSource();
            this.setStoreSessionStatus(-1);
            return;
          }
        },
        onmessage: e => {
          if (e && e.data) {
            let data = JSON.parse(e.data);
            console.log('===>', new Date().getTime(), data);
            this.sseResponse = data;
            //待替换的数据，需要前端组装
            let commonData = {
              ...data,
              ...this.sseParams,
              query: prompt,
              fileList: this.fileList,
              response: '',
              filepath: data.file_url || '',
              requestFileUrls: this.queryFilePath
                ? [this.queryFilePath]
                : data.requestFileUrls,
              searchList: data.search_list || [],
              gen_file_url_list: data.gen_file_url_list || [],
              thinkText: i18n.t('agent.thinking'),
              toolText: '使用工具中...',
              isOpen: true,
              showScrollBtn: null,
              citations: [],
              subConversions: [], // 初始化子会话列表
              messageSequence: [], // 初始化消息序列，用于平铺渲染
              _lastOrder: -1, // 内部追踪最后一次的 order
              responseFiles: [], // 此处传空，统一通过 this.responseFiles 获取
            };

            // 实时同步并处理 responseFiles
            if (data.responseFiles && data.responseFiles.length) {
              this.responseFiles = data.responseFiles.map(r =>
                transformSkillData(r),
              );
            }

            if (data.code === 0) {
              // 处理子会话消息 (eventType === 0)
              if (data.eventType === 1 && data.eventData) {
                const { id, name, status, timeCost, profile } = data.eventData;
                let subConversion = this._subConversionsMap.get(id);
                let subProcessor = this._subConversionProcessors.get(id);

                if (!subConversion) {
                  subConversion = {
                    id,
                    name,
                    status, // 1开始、2输出中、3结束、4处理失败
                    timeCost,
                    profile, //头像
                    response: '',
                    stableChunks: [],
                    activeResponse: '',
                    isOpen: false, // 默认收起
                    searchList: data.search_list || [], // 初始化 searchList
                    citationsTagList: [], // 已引用的出处索引
                  };
                  this._subConversionsMap.set(id, subConversion);

                  // 初始化流处理器
                  subProcessor = new StreamProcessor({
                    lastIndex,
                    md,
                    parseSub: (text, index, searchList) =>
                      parseSubConversation(text, index, searchList, id),
                    convertLatexSyntax,
                    searchList: subConversion.searchList,
                  });
                  this._subConversionProcessors.set(id, subProcessor);
                } else {
                  // 更新状态和耗时
                  subConversion.status = status;
                  if (timeCost) subConversion.timeCost = timeCost;
                  // 如果后续包中有 search_list，则更新
                  if (data.search_list && data.search_list.length) {
                    subConversion.searchList = data.search_list;
                    subProcessor.updateSearchList(data.search_list);
                  }
                }

                // 累加回复内容并处理流
                if (data.response) {
                  // 处理转义换行符
                  let processedResponse = data.response.replace(/\\n/g, '\n');
                  subConversion.response += processedResponse;
                  subProcessor.append(processedResponse);
                  const renderResult = subProcessor.getRenderResult();
                  subConversion.stableChunks = renderResult.stableChunks;
                  subConversion.activeResponse = renderResult.activeResponse;
                  // StreamProcessor 增量维护的引文列表
                  subConversion.citationsTagList = renderResult.citations || [];
                }

                // 更新消息序列
                let sequence =
                  sessionCom.getSessionData()['history'][lastIndex]
                    .messageSequence || [];
                if (data.order !== undefined && data.order !== null) {
                  let currentSubItem = sequence.find(
                    item =>
                      item.type === 'sub' &&
                      item.id === id &&
                      item.order === data.order,
                  );
                  if (!currentSubItem) {
                    currentSubItem = {
                      type: 'sub',
                      id: id,
                      order: data.order,
                    };
                    sequence.push(currentSubItem);
                  }
                }

                // 构造 fillData
                // 获取最新的子会话列表
                const subConversionsList = Array.from(
                  this._subConversionsMap.values(),
                );

                let fillData = {
                  ...commonData,
                  finish:
                    this._currentMainFinish !== undefined
                      ? this._currentMainFinish
                      : 0,
                  subConversions: subConversionsList,
                  messageSequence: sequence,
                };

                sessionCom.replaceLastData(lastIndex, fillData);
                // 如果子智能体结束或失败，可能需要滚动到底部
                if (status === 3 || status === 4) {
                  this.$nextTick(() => sessionCom.scrollBottom());
                }
              } else {
                // 主智能体消息 (eventType === 0 或 undefined)
                // 更新当前主智能体 finish 状态
                this._currentMainFinish = data.finish;

                // 根据 order 获取或创建对应的 processor
                const currentOrder = data.order !== undefined ? data.order : 0;
                let mainProcessor = this._mainProcessors.get(currentOrder);

                if (!mainProcessor) {
                  mainProcessor = new StreamProcessor({
                    lastIndex,
                    md,
                    parseSub,
                    convertLatexSyntax,
                  });
                  this._mainProcessors.set(currentOrder, mainProcessor);
                }

                //finish 0：进行中  1：关闭   2:敏感词关闭
                let _sentence = data.response;
                this._print.print(
                  {
                    response: _sentence,
                    finish: data.finish,
                  },
                  commonData,
                  (worldObj, search_list) => {
                    this.setStoreSessionStatus(0);
                    mainProcessor.updateSearchList(search_list);
                    mainProcessor.append(worldObj.world);

                    const renderResult = mainProcessor.getRenderResult();

                    // 更新消息序列
                    let sequence =
                      sessionCom.getSessionData()['history'][lastIndex]
                        .messageSequence || [];

                    if (data.order !== undefined && data.order !== null) {
                      let currentMainItem = sequence.find(
                        item =>
                          item.type === 'main' && item.order === data.order,
                      );

                      if (!currentMainItem) {
                        currentMainItem = {
                          type: 'main',
                          order: data.order,
                          renderedContent: '',
                          stableChunks: [],
                          activeResponse: '',
                        };
                        sequence.push(currentMainItem);
                      }

                      currentMainItem.renderedContent = renderResult.response;
                      currentMainItem.stableChunks = renderResult.stableChunks;
                      currentMainItem.activeResponse =
                        renderResult.activeResponse;
                    }

                    // 获取最新的子会话列表
                    const subConversionsList = Array.from(
                      this._subConversionsMap.values(),
                    );

                    let fillData = {
                      ...commonData,
                      ...renderResult,
                      finish: worldObj.finish,
                      searchList:
                        search_list && search_list.length
                          ? search_list.map(n => ({
                              ...n,
                              snippet: md.render(n.snippet),
                            }))
                          : [],
                      subConversions: subConversionsList,
                      messageSequence: sequence,
                      responseFiles: JSON.parse(
                        JSON.stringify(this.responseFiles),
                      ),
                    };
                    sessionCom.replaceLastData(lastIndex, fillData);
                    if (worldObj.finish !== 0) {
                      if (worldObj.finish === 4) {
                        let fillData = {
                          ...commonData,
                          response: i18n.t('sse.sensitiveTips'),
                          subConversions: subConversionsList,
                          messageSequence: sequence,
                          responseFiles: JSON.parse(
                            JSON.stringify(this.responseFiles),
                          ),
                        };
                        sessionCom.replaceLastData(lastIndex, fillData);
                        this.$nextTick(() => {
                          sessionCom.scrollBottom();
                        });
                      }
                      this.setStoreSessionStatus(-1);
                    }

                    if (worldObj.isEnd && worldObj.finish === 1) {
                      this.setStoreSessionStatus(-1);
                      this._currentMainFinish = undefined;
                    }
                  },
                );
              }
            } else if (data.code === 7 || data.code === -1 || data.code === 1) {
              this.setStoreSessionStatus(-1);
              // 获取最新的子会话列表，防止被覆盖
              const subConversionsList = this._subConversionsMap
                ? Array.from(this._subConversionsMap.values())
                : [];
              let fillData = {
                ...commonData,
                response: data.message,
                subConversions: subConversionsList,
                responseFiles: JSON.parse(JSON.stringify(this.responseFiles)),
              };
              sessionCom.replaceLastData(lastIndex, fillData);
              this._currentMainFinish = undefined;
            }
          }
        },
      });
    },
  },
};
