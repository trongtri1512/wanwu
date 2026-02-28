<template>
  <div class="compare-container compare-container-wrapper page-wrapper">
    <div class="compare-header">
      <div class="compare-header-left">
        <span class="el-icon-arrow-left go-back" @click="goBack"></span>
        <h3>{{ $t('tempSquare.promptCompare') }}</h3>
      </div>
      <el-button type="primary" size="small" @click="addPromptField">
        <i class="el-icon-plus" style="margin-right: 4px"></i>
        {{ $t('tempSquare.addPrompt') }}
      </el-button>
    </div>
    <div class="compare-content">
      <div class="compare-middle">
        <div class="prompt-field-list">
          <div
            class="prompt-field-item"
            :style="{
              width: `calc((100% - ${(promptFields.length - 1) * 10}px) / ${
                promptFields.length
              })`,
            }"
            v-for="(field, index) in promptFields"
            :key="field.id"
          >
            <PromptField
              :ref="`promptField${index}`"
              :fieldIndex="index"
              :editForm="editForm"
              @closePrompt="closePrompt"
              :isSelected="selectedFieldIndex === index"
              @selectField="handleFieldSelect"
            />
          </div>
        </div>
      </div>
      <div class="compare-bottom">
        <streamInputField
          ref="editable"
          source="promptCompare"
          :visibleClearHistory="false"
          :hasHistory="false"
          :fileTypeArr="fileTypeArr"
          type="compare"
          @preSend="handlePromptSubmit"
        />
      </div>
    </div>
  </div>
</template>

<script>
import PromptField from './promptField.vue';
// import EditableDivV3 from '../EditableDivV3';
import streamInputField from '@/components/stream/streamInputField';
import { getAgentInfo } from '@/api/agent';

export default {
  name: 'PromptCompare',
  provide() {
    return {
      getEditableRef: () => this.$refs.editable,
    };
  },
  components: {
    PromptField,
    streamInputField,
    // EditableDivV3,
  },
  data() {
    return {
      promptFields: [{ id: Date.now() + Math.random() }],
      fileTypeArr: [],
      currentModel: null,
      editForm: null,
      selectedFieldIndex: -1,
    };
  },
  created() {
    this.getAgentInfo();
  },
  methods: {
    handleFieldSelect(index) {
      this.selectedFieldIndex = index;
    },
    handlePromptSubmit() {
      const editable = this.$refs.editable;
      if (!editable || typeof editable.getPrompt !== 'function') return;
      const promptText = (editable.getPrompt() || '').trim();
      if (!promptText) return;
      //验证提示词对比是否为空
      if (!this.promptValidate()) {
        return;
      }
      this.$nextTick(() => {
        for (let i = 0; i < this.promptFields.length; i += 1) {
          const refName = `promptField${i}`;
          const fieldRef = this.$refs[refName];
          const component = Array.isArray(fieldRef) ? fieldRef[0] : fieldRef;
          if (component && typeof component.preSend === 'function') {
            component.preSend(promptText);
          }
        }
      });
    },
    promptValidate() {
      //验证提示词对比是否为空
      const emptyFields = [];
      for (let i = 0; i < this.promptFields.length; i += 1) {
        const refName = `promptField${i}`;
        const fieldRef = this.$refs[refName];
        const component = Array.isArray(fieldRef) ? fieldRef[0] : fieldRef;
        if (component) {
          const systemPrompt = (component.systemPrompt || '').trim();
          if (!systemPrompt) {
            emptyFields.push(
              i === 0
                ? this.$t('agent.form.systemPrompt')
                : `${this.$t('tempSquare.comparePrompt')}${i}`,
            );
          }
        }
      }
      if (emptyFields.length > 0) {
        const fieldNames = emptyFields.join('、');
        this.$message.warning(
          `${fieldNames}${this.$t('tempSquare.promptRules')}`,
        );
        return false;
      }
      return true;
    },
    getAgentInfo() {
      getAgentInfo({ assistantId: this.$route.params.id }).then(res => {
        if (res.code === 0) {
          const detail = res.data || {};
          const recommendQuestion = Array.isArray(detail.recommendQuestion)
            ? detail.recommendQuestion.filter(item => item && item.value)
            : [];
          this.editForm = {
            ...detail,
            recommendQuestion,
          };
        }
      });
    },
    closePrompt(index) {
      this.promptFields.splice(index, 1);
    },
    addPromptField() {
      if (this.promptFields.length > 4) {
        this.$message.warning(this.$t('tempSquare.promptCompareLimit'));
        return;
      }
      this.promptFields.push({ id: Date.now() + Math.random() });
    },
    goBack() {
      const id = this.$route.params.id;
      this.$router.push({ path: `/agent/test/?id=${id}` });
    },
  },
};
</script>

<style scoped lang="scss">
.compare-container-wrapper {
  height: calc(100vh - 32px) !important;
  overflow: hidden;
}

.compare-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  padding: 0 10px;
  box-sizing: border-box;
}

.compare-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 10px 14px 10px;
  border-bottom: 1px solid #eaeaea;
  .compare-header-left {
    display: flex;
    align-items: center;
    .go-back {
      font-size: 18px;
      cursor: pointer;
    }
    h3 {
      font-size: 18px;
      font-weight: 800;
      color: #434c6c;
      margin-left: 10px;
    }
  }
}
.compare-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  .compare-middle {
    flex: 1;
    margin: 10px 10px 0 10px;
    overflow-y: auto;
    .prompt-field-list {
      display: flex;
      height: 100%;
      gap: 10px;
      overflow-x: auto;
      .prompt-field-item {
        flex: 1;
        height: 100%;
      }
    }
  }
  .compare-bottom {
    padding: 10px;
  }
}
</style>
