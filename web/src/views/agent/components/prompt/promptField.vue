<template>
  <div class="compare-container">
    <div class="compare-top">
      <div class="drawer-info">
        <div class="promptTitle">
          <div style="display: flex; align-items: center">
            <h3>
              {{
                fieldIndex === 0
                  ? $t('agent.form.systemPrompt')
                  : $t('tempSquare.comparePrompt')
              }}
            </h3>
            <el-button
              :type="isSelected ? 'primary' : 'default'"
              size="mini"
              @click="handleSelectPrompt"
            >
              <span v-if="!isSelected">
                <span>{{ $t('tempSquare.select') }}</span>
              </span>
              <span v-else>
                <i class="el-icon-check" style="margin-right: 4px"></i>
                <span>{{ $t('tempSquare.selected') }}</span>
              </span>
            </el-button>
          </div>
          <div class="prompt-title-icon">
            <el-tooltip
              class="item"
              effect="dark"
              :content="$t('agent.form.submitToPrompt')"
              placement="top-start"
            >
              <span
                class="el-icon-folder-add tool-icon"
                @click="handleShowPrompt"
              ></span>
            </el-tooltip>
            <el-tooltip
              class="item"
              effect="dark"
              :content="$t('tempSquare.promptOptimize')"
              placement="top-start"
            >
              <span
                class="el-icon-s-help tool-icon"
                @click="showPromptOptimize"
              ></span>
            </el-tooltip>
            <el-tooltip
              class="item"
              effect="dark"
              :content="$t('tempSquare.closePrompt')"
              placement="top-start"
              v-if="fieldIndex > 0"
            >
              <span
                class="el-icon-close tool-icon"
                @click="handleClosePrompt"
              ></span>
            </el-tooltip>
          </div>
        </div>
        <div class="rl prompt-input">
          <el-input
            class="desc-input"
            v-model="systemPrompt"
            :placeholder="$t('agent.form.promptTips')"
            type="textarea"
            show-word-limit
            :rows="4"
          ></el-input>
        </div>
      </div>
    </div>
    <div class="compare-bottom">
      <div class="compare-bottom-content">
        <div v-show="echo" class="session rl echo">
          <streamGreetingField
            :editForm="editForm"
            :sessionItemWidth="sessionItemWidth"
            :showRecommendQuestion="showRecommendQuestion"
          />
        </div>
        <!--对话-->
        <div v-show="!echo" class="center-session">
          <streamMessageField
            ref="sessionComLocal"
            class="component"
            :sessionStatus="instanceSessionStatus"
            @clearHistory="clearHistory"
            @queryCopy="queryCopy"
            @refresh="refresh"
            :type="type"
            :defaultUrl="editForm && editForm.avatar && editForm.avatar.path"
          />
        </div>
      </div>
    </div>
    <!-- 提示词优化 -->
    <PromptOptimize ref="promptOptimize" @promptSubmit="promptSubmit" />
    <!-- 提交至提示词 -->
    <createPrompt :isCustom="true" :type="promptType" ref="createPrompt" />
  </div>
</template>

<script>
import streamGreetingField from '@/components/stream/streamGreetingField';
import streamMessageField from '@/components/stream/streamMessageField';
import PromptOptimize from '@/components/promptOptimize.vue';
import createPrompt from '@/components/createApp/createPrompt.vue';
import sseMethodMixin from '@/mixins/sseMethod';
export default {
  name: 'PromptCompareField',
  mixins: [sseMethodMixin],
  inject: {
    getEditableRef: {
      default: () => null,
    },
  },
  props: {
    fieldIndex: {
      type: Number,
      default: 0,
    },
    editForm: {
      typeof: Object,
      default: null,
    },
    editForm: {
      type: Object,
      default: null,
    },
    isSelected: {
      type: Boolean,
      default: false,
    },
  },
  components: {
    streamGreetingField,
    streamMessageField,
    PromptOptimize,
    createPrompt,
  },
  watch: {
    fieldIndex: {
      handler(newVal) {
        if (newVal === 0 && this.editForm && this.editForm.instructions) {
          this.systemPrompt = this.editForm.instructions;
        }
      },
      immediate: true,
    },
    editForm: {
      handler(newVal) {
        if (this.fieldIndex === 0 && newVal && newVal.instructions) {
          this.systemPrompt = newVal.instructions;
        }
      },
      immediate: true,
      deep: true,
    },
  },
  data() {
    return {
      showRecommendQuestion: false,
      promptType: 'create',
      sessionItemWidth: '19vw',
      systemPrompt: '',
      echo: true,
      type: 'agentChat',
      fieldId: 'prompt-field-' + this._uid, //区分提示词对比每个字段的会话id
    };
  },
  mounted() {
    const currentSession = this.$refs.sessionComLocal;
    if (currentSession) {
      this.$refs['session-com'] = currentSession;
    }
  },
  methods: {
    preSend(promptText, fileList, file_info) {
      let sessionCom = this.$refs.sessionComLocal;
      if (!sessionCom || typeof sessionCom.getList !== 'function') return;

      let historyList = sessionCom.getList();
      let lastIndex = Array.isArray(historyList) ? historyList.length : 0;
      this.fileList =
        fileList ||
        (this.getEditableRef && this.getEditableRef().getFileList());
      let fileInfo =
        (this.getEditableRef && this.getEditableRef().getFileIdList()) ||
        file_info;
      this.isTestChat = true;
      this.setSseParams({
        assistantId: this.editForm && this.editForm.assistantId,
        conversationId: '',
        fileInfo: fileInfo,
        systemPrompt: this.systemPrompt,
        isCompare: true, // 提示词对比true以避免影响草稿态的智能体历史上下文
      });
      this.echo = false;
      this.sendEventSource(promptText, '', lastIndex);
    },
    clearHistory() {
      this.stopEventSource();
      if (
        this.$refs.sessionComLocal &&
        typeof this.$refs.sessionComLocal.clearData === 'function'
      ) {
        this.$refs.sessionComLocal.clearData();
      }
    },
    handleShowPrompt() {
      this.$refs.createPrompt.openDialog({ prompt: this.systemPrompt });
    },
    showPromptOptimize() {
      if (!this.systemPrompt) {
        this.$message.warning(this.$t('tempSquare.promptOptimizeHint'));
        return;
      }
      this.$refs.promptOptimize.openDialog({ prompt: this.systemPrompt });
    },
    handleClosePrompt() {
      this.$emit('closePrompt', this.fieldIndex);
    },
    promptSubmit(prompt) {
      this.systemPrompt = prompt;
    },
    handleSelectPrompt() {
      this.$emit('selectField', this.fieldIndex);
      this.copy(this.systemPrompt) && this.copycb();
    },
    copy(text) {
      text = text.replaceAll('<br/>', '\n');
      let textareaEl = document.createElement('textarea');
      textareaEl.setAttribute('readonly', 'readonly');
      textareaEl.value = text;
      document.body.appendChild(textareaEl);
      textareaEl.select();
      let res = document.execCommand('copy');
      document.body.removeChild(textareaEl);
      return res;
    },
    copycb() {
      this.$message.success(this.$t('agent.copyTips'));
    },
  },
};
</script>

<style scoped lang="scss">
.compare-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  box-sizing: border-box;
  background: #f2f7ff8f;
  border: 1px solid #eaeaea;
  border-radius: 8px;
  min-height: 0; //防止容器高度塌陷
  overflow: hidden;
}
.compare-container:hover {
  border: 1px solid $color;
}

.compare-top {
  flex: 2;
}
.compare-bottom {
  flex: 8;
  position: relative;
  overflow: hidden;
}
.compare-bottom-title {
  font-size: 16px;
  font-weight: 600;
  padding: 10px;
}

.compare-bottom-content {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0 10px;
  box-sizing: border-box;
}

.compare-bottom-content .session,
.compare-bottom-content {
  height: 100%;
  overflow: hidden;
}

.compare-bottom-content .center-session {
  height: 100%;
  padding-bottom: 10px;
}

.drawer-info {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.promptTitle {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0 0;
  h3 {
    font-size: 16px;
    font-weight: 800;
    margin-right: 6px;
  }
  ::v-deep.el-button--mini,
  .el-button--mini.is-round {
    font-size: 12px;
    height: 24px;
    padding: 0 10px;
  }
}

.prompt-title-icon {
  display: flex;
  align-items: center;
  span {
    font-size: 16px;
    color: #5c6ac4;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    margin-left: 10px;
  }
  .tool-icon {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: #e0e7ff;
    color: $color;
    img {
      width: 16px;
      height: 16px;
    }
  }
}

.prompt-input {
  padding: 10px 0;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.desc-input ::v-deep .el-textarea__inner {
  background-color: transparent !important;
  border: 1px solid #d3d7dd !important;
  padding: 15px;
}
</style>
