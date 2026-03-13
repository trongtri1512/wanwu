<template>
  <div class="full-content flex">
    <el-main class="scroll">
      <div class="smart-center" style="padding: 0">
        <!--开场白设置-->
        <div v-if="echo" class="session rl echo">
          <!-- <streamGreetingField
            :editForm="editForm"
            sessionItemWidth="100%"
            @setProloguePrompt="setProloguePrompt"
          /> -->
          <h2 class="slogan">{{ $t('tempSquare.skills.createSlogan') }}</h2>
          <InputAreaField
            ref="editable"
            source="perfectReminder"
            :fileTypeArr="fileTypeArr"
            :type="type"
            :hasHistory="hasHistory"
            :landing="true"
            :visibleClearHistory="false"
            minHeight="120px"
            :modelConfig="sharedModelConfig"
            @preSend="preSend"
            @setSessionStatus="setSessionStatus"
            @clearHistory="handleClearHistory"
            @inputHeightChange="handleInputHeightChange"
          />
        </div>
        <!--对话-->
        <div v-show="!echo" class="center-session">
          <streamMessageField
            ref="session-com"
            class="component"
            :chatType="'agent'"
            :sessionStatus="sessionStatus"
            :recommendConfig="recommendConfig"
            :supportClear="false"
            @clearHistory="clearHistory"
            @refresh="refresh"
            @queryCopy="queryCopy"
            @handleRecommendClick="handleRecommendClick"
            :modelIconUrl="AgentIcon"
          >
            <template #afterContent="{ skillsList }">
              <div class="skills-card-list">
                <SkillCard
                  v-for="skill in skillsList"
                  :key="skill.skillId"
                  :info="skill"
                  :type="3"
                  @download="handleDownload"
                  @sendToResource="handleSendToResource"
                />
              </div>
            </template>
          </streamMessageField>
        </div>
        <!--停止生成-重新生成-->
        <div class="center-editable">
          <div v-show="stopBtShow" class="stop-box">
            <span v-show="sessionStatus === 0" class="stop" @click="preStop">
              <img
                class="stop-icon mdl"
                :src="require('@/assets/imgs/stop.png')"
              />
              <span class="mdl">{{ $t('agent.stop') }}</span>
            </span>
            <span v-show="sessionStatus !== 0" class="stop" @click="refresh">
              <img
                class="stop-icon mdl"
                :src="require('@/assets/imgs/refresh.png')"
              />
              <span class="mdl">{{ $t('agent.refresh') }}</span>
            </span>
          </div>
          <!-- 输入框 -->
          <InputAreaField
            v-if="!echo"
            ref="editable"
            class="raw-inputAreaField"
            source="perfectReminder"
            :fileTypeArr="fileTypeArr"
            :type="type"
            :hasHistory="hasHistory"
            :landing="true"
            :visibleClearHistory="true"
            minHeight="22px"
            :modelConfig="sharedModelConfig"
            @preSend="preSend"
            @setSessionStatus="setSessionStatus"
            @clearHistory="handleClearHistory"
            @inputHeightChange="handleInputHeightChange"
          />
          <!-- 版权信息 -->
          <div v-if="appUrlInfo" class="appUrlInfo">
            <span v-if="appUrlInfo.copyrightEnable">
              {{ $t('app.copyright') }}: {{ appUrlInfo.copyright }}
            </span>
            <span v-if="appUrlInfo.privacyPolicyEnable">
              {{ $t('app.privacyPolicy') }}:
              <a
                :href="appUrlInfo.privacyPolicy"
                target="_blank"
                style="color: var(--color)"
              >
                {{ appUrlInfo.privacyPolicy }}
              </a>
            </span>
            <span v-if="appUrlInfo.disclaimerEnable">
              {{ $t('app.disclaimer') }}: {{ appUrlInfo.disclaimer }}
            </span>
          </div>
        </div>
      </div>
    </el-main>
  </div>
</template>

<script>
import streamMessageField from '@/components/stream/streamMessageField';
import streamInputField from '@/components/stream/streamInputField';
import streamGreetingField from '@/components/stream/streamGreetingField';
import SkillCard from '@/views/templateSquare/skills/card.vue';
import InputAreaField from './InputAreaField';
import {
  parseSub,
  convertLatexSyntax,
  parseSubConversation,
} from '@/utils/util.js';
import {
  createCustomSkillConversation,
  delCustomSkillConversation,
  getCustomSkillConversationDetail,
  sendCustomSkillToResource,
  clearSkillConversation,
} from '@/api/templateSquare';
import sseMethod from '@/mixins/sseMethod';
import { md } from '@/mixins/markdown-it';
import { mapGetters, mapState } from 'vuex';
import AgentIcon from '@/assets/imgs/agent.svg';
import { directDownload } from '@/utils/util';

export default {
  inject: {
    getHeaderConfig: {
      default: () => null,
    },
  },
  props: {
    editForm: {
      type: Object,
      default: null,
    },
    chatType: {
      type: String,
      default: '',
    },
    type: {
      type: String,
      default: 'agentChat',
    },
    appUrlInfo: {
      type: Object,
      default: null,
    },
  },
  components: {
    streamMessageField,
    streamInputField,
    streamGreetingField,
    InputAreaField,
    SkillCard,
  },
  mixins: [sseMethod],
  computed: {
    ...mapGetters('app', ['sessionStatus']),
    ...mapGetters('menu', ['basicInfo']),
    ...mapGetters('user', ['commonInfo']),
    ...mapState('user', ['userInfo']),
    hasHistory() {
      return !this.echo;
    },
  },
  data() {
    return {
      echo: true,
      fileTypeArr: ['doc/*', 'image/*'],
      hasDrawer: false,
      drawer: true,
      fileId: [],
      recommendConfig: {
        reqController: new AbortController(),
        list: [],
        loading: false,
      },
      recommendTimer: null,
      sharedModelConfig: {
        modelId: '',
      },
      AgentIcon,
    };
  },
  methods: {
    createConversion() {
      if (this.echo) {
        this.$message({
          type: 'info',
          message: this.$t('app.switchSession'),
          customClass: 'dark-message',
          iconClass: 'none',
          duration: 1500,
        });
        return;
      }
      this.conversationId = '';
      this.echo = true;
      this.clearPageHistory();
      this.$emit('setHistoryStatus');
    },
    //切换对话
    conversationClick(n) {
      if (this.sessionStatus === 0) {
        return;
      } else {
        this.stopBtShow = false;
      }

      this.$emit('setHistoryStatus');
      this.amswerNum = 0;
      n.active = true;
      this.clearPageHistory();
      this.echo = false;
      this.conversationId = n.conversationId;
      this.getConversationDetail(this.conversationId, true);
    },
    async getConversationDetail(id, loading) {
      loading && this.$refs['session-com'].doLoading();
      const res = await getCustomSkillConversationDetail({
        conversationId: id,
        pageSize: 1000,
        pageNo: 1,
      });

      if (res.code === 0) {
        let history = this.convertHistoryData(res.data.list);

        this.$refs['session-com'].replaceHistory(history);
      }
    },
    //删除对话
    async preDelConversation(n) {
      if (this.sessionStatus === 0) {
        return;
      }
      let res = await delCustomSkillConversation({
        conversationId: n.conversationId,
      });
      if (res.code === 0) {
        this.$emit('reloadList');
        if (this.conversationId === n.conversationId) {
          this.conversationId = '';
          this.$refs['session-com'].clearData();
        }
        this.echo = true;
      }
    },
    /*------会话------*/
    async preSend(val, fileList, fileInfo, modelConfig) {
      if (this.recommendTimer) {
        clearInterval(this.recommendTimer);
        this.recommendTimer = null;
      }
      if (this.recommendConfig.loading) {
        this.recommendConfig.reqController.abort();
        this.recommendConfig.reqController = new AbortController();
      }
      this.recommendConfig.list = [];
      this.recommendConfig.loading = false;
      this.inputVal = val || this.$refs['editable'].getPrompt();
      this.fileId = fileInfo || [];
      this.isTestChat = this.chatType === 'test';
      this.fileList = fileList || this.$refs['editable'].getFileList();
      if (!this.inputVal) {
        this.$message.warning(this.$t('agent.inputContent'));
        return;
      }
      if (!modelConfig.modelId) {
        this.$message.warning(this.$t('tempSquare.skills.noModelIdTips'));
        return;
      }
      //如果是新会话，先创建
      if (!this.conversationId && this.chatType === 'chat') {
        const res = await createCustomSkillConversation({
          title: this.inputVal,
          // assistantId: this.editForm.assistantId,
        });

        if (res.code === 0) {
          this.conversationId = res.data.conversationId;
          this.$emit('reloadList', true);
          this.setParams(modelConfig);
        }
      } else {
        this.setParams(modelConfig);
      }
    },
    setParams(modelConfig) {
      const fileInfo = this.$refs['editable'].getFileIdList();
      let fileId = !fileInfo.length ? this.fileId : fileInfo;
      this.setSseParams({
        conversationId: this.conversationId,
        fileInfo: fileId,
        assistantId: '', // skills无assistantId
        modelConfig,
      });
      this.doSkillsSend();
      this.echo = false;
    },
    /*--右侧提示词--*/
    showDrawer() {
      this.drawer = true;
    },
    hideDrawer() {
      this.drawer = false;
    },
    async getReminderList(cb) {
      let res = await getTemplateList({ pageNo: 0, pageSize: 0, title: '' });
      if (res.code === 0) {
        this.reminderList = res.data.list || [];
        cb && cb();
      }
    },
    reminderClick(n) {
      this.$refs['editable'].setPrompt(n.prompt);
    },
    // 打印结束回调
    onMainPrintEnd() {
      const history = this.$refs['session-com'].getSessionData().history;
      const lastMessage = history[history.length - 1];

      // 只有当最后一条消息存在且 finish 状态为 1 (真正结束) 时才触发推荐
      if (
        lastMessage &&
        lastMessage.finish === 1 &&
        this.editForm.recommendConfig &&
        this.editForm.recommendConfig.recommendEnable &&
        this.editForm.recommendConfig.modelConfig.modelId
      ) {
        this.recommendConfig.list = [];
      }
    },
    handleRecommendClick(val) {
      this.preSend(val);
    },
    // 转换智能体历史记录数据
    convertHistoryData(data) {
      return data
        ? data.map((n, index) => {
            const sequence = [];
            let fullResponse = '';

            // 处理主智能体片段 (responseList)
            if (n.responseList && n.responseList.length) {
              n.responseList.forEach(item => {
                fullResponse += item.response || '';

                sequence.push({
                  type: 'main',
                  order: item.order,
                  renderedContent: md.render(
                    parseSub(convertLatexSyntax(item.response), index),
                  ),
                });
              });
            } else if (n.response) {
              // 处理非分段片段
              fullResponse = n.response;
            }

            // 处理子会话片段 (subConversationList)
            const subConversions = n.subConversationList
              ? n.subConversationList.map(m => {
                  const citationsTagList = (
                    (m.response || '').match(/\【([0-9]{0,2})\^\】/g) || []
                  ).map(item => Number(item.match(/\【([0-9]{0,2})\^\】/)[1]));

                  const processedSub = {
                    ...m,
                    citationsTagList,
                    searchList:
                      typeof m.searchList === 'string'
                        ? JSON.parse(m.searchList || '[]')
                        : m.searchList || [],
                    response: md.render(
                      parseSubConversation(
                        convertLatexSyntax(m.response || ''),
                        index,
                        m.searchList,
                        m.id,
                      ),
                    ),
                  };

                  sequence.push({
                    type: 'sub',
                    id: m.id,
                    order: m.order,
                  });

                  return processedSub;
                })
              : [];

            // 根据 order 排序
            sequence.sort((a, b) => (a.order || 0) - (b.order || 0));
            const r = {
              ...n,
              query: n.prompt,
              finish: 1, //兼容流式问答
              response: md.render(
                parseSub(convertLatexSyntax(fullResponse), index),
              ),
              oriResponse: fullResponse,
              searchList:
                typeof n.searchList === 'string'
                  ? JSON.parse(n.searchList || '[]')
                  : n.searchList || [],
              fileList: n.requestFiles,
              gen_file_url_list: n.responseFileUrls || [],
              subConversions: subConversions,
              messageSequence: sequence,
              isOpen: true,
              toolText: this.$t('agent.tooled'),
              thinkText: this.$t('agent.thinked'),
              showScrollBtn: null,
              responseFiles: n.responseFiles
                ? n.responseFiles.map(r => this.transformSkillData(r))
                : [],
            };
            return r;
          })
        : [];
    },
    // 清空会话
    async handleClearHistory() {
      const history = this.$refs['session-com'].session_data.history;
      if (!history || !history.length) return;
      this.handleClearSkillConversation();
    },
    // 清空skill会话
    async handleClearSkillConversation() {
      try {
        const res = await clearSkillConversation({
          conversationId: this.conversationId,
        });
        if (res.code === 0) {
          this.clearHistory();
        }
      } catch (error) {
        throw error;
      }
    },
    // 处理输入框高度变化
    handleInputHeightChange(height) {
      this.$refs['session-com'] &&
        this.$refs['session-com'].setHistoryBoxHeight(height);
    },
    // 下载skill文件
    handleDownload(fileInfo) {
      const { fileUrl } = fileInfo;
      directDownload(fileUrl);
    },
    // 发布skill到资源库
    async handleSendToResource(fileInfo) {
      const params = {
        conversationId: this.conversationId,
        skillSaveId: fileInfo.skillSaveId,
      };
      try {
        const res = await sendCustomSkillToResource(params);
        if (res.code === 0) {
          this.$message.success(this.$t('common.info.send'));
        }
      } catch (error) {
        throw error;
      }
    },
    // 转换流式数据的skill文件结构
    transformSkillData(rawData) {
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
    },
  },
  mounted() {},
  beforeDestroy() {
    if (this.recommendTimer) {
      clearInterval(this.recommendTimer);
      this.recommendTimer = null;
    }
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/chat.scss';
.appUrlInfo {
  margin-top: 10px;
  display: flex;
  justify-content: center;
  span {
    cursor: pointer;
    color: #bbb;
    margin-right: 15px;
  }
}

.echo {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-direction: column;
  gap: 32px;
  .slogan {
    font-size: 4rem;
  }
}

.skills-card-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 12px;
  .card {
    min-width: 270px;
  }
}
::v-deep .raw-inputAreaField {
  .editable-box .echo-img-box {
    bottom: unset;
    top: -70px;
  }
}
</style>
