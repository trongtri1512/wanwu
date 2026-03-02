<template>
  <div class="qa-database-container">
    <div class="qa-database-header">
      <div class="header-left">
        <img
          :src="require('@/assets/imgs/require.png')"
          class="required-label"
          v-if="required"
        />
        <span class="header-title">
          {{ labelText }}
        </span>
      </div>
      <div class="header-right">
        <span class="common-add" @click="handleAdd">
          <span class="el-icon-plus"></span>
          <span class="handleBtn">{{ $t('knowledgeSelect.add') }}</span>
        </span>
        <span class="common-add" @click="showknowledgeRecallSet">
          <el-tooltip
            class="item"
            effect="dark"
            :content="$t('searchConfig.title')"
            placement="top-start"
          >
            <span class="el-icon-s-operation operation">
              <span class="handleBtn">{{ $t('agent.form.config') }}</span>
            </span>
          </el-tooltip>
        </span>
      </div>
    </div>
    <div class="qa-database-content">
      <div
        class="action-list"
        v-if="showKnowledgeList"
        :class="{
          'single-row': appType === 'agent',
          'two-row': appType !== 'agent',
        }"
      >
        <div
          v-for="(item, index) in knowledgeList"
          :key="item.id"
          class="action-item"
        >
          <div
            class="name"
            @click="handleKnowledgeLink(item.category, item.id)"
          >
            <span>
              {{ item.name }}
            </span>
          </div>
          <div class="bt">
            <el-tooltip
              v-if="item.external !== 1"
              class="item"
              effect="dark"
              :content="$t('agent.form.metaDataFilter')"
              placement="top-start"
            >
              <span
                class="el-icon-setting del"
                @click="handleSetting(item, index)"
                style="margin-right: 10px"
              ></span>
            </el-tooltip>
            <span
              class="el-icon-delete del"
              @click="handleDelete(index)"
            ></span>
          </div>
        </div>
      </div>
    </div>
    <knowledgeSelect
      ref="knowledgeSelect"
      :category="category"
      @getKnowledgeData="getKnowledgeData"
    />
    <metaDataFilterField
      ref="metaDataFilterField"
      :knowledgeId="currentKnowledgeId"
      :metaData="currentMetaData"
      @submitMetaData="submitMetaData"
      :category="category"
    />
    <knowledgeRecallField
      ref="knowledgeRecallField"
      :showGraphSwitch="showGraphSwitch"
      @setKnowledgeSet="knowledgeRecallSet"
      :config="knowledgeRecallConfig"
      :category="category"
      :knowledgeCategory="knowledgeCategory"
      :isAllExternal="isAllExternalKnowledgeSelected"
    />
  </div>
</template>
<script>
import knowledgeSelect from '@/components/knowledgeSelect.vue';
import metaDataFilterField from './metaDataFilterField.vue';
import knowledgeRecallField from './knowledgeRecallField.vue';
import { KNOWLEDGE } from '@/views/knowledge/constants';
export default {
  name: 'QaDatabase',
  components: { knowledgeSelect, metaDataFilterField, knowledgeRecallField },
  props: {
    knowledgeConfig: {
      type: Object,
      default: () => null,
      require: true,
    },
    category: {
      type: Number,
      default: KNOWLEDGE,
    },
    knowledgeCategory: {
      type: Number,
      default: KNOWLEDGE,
    },
    required: {
      type: Boolean,
      default: false,
    },
    searchConfig: {
      type: Object,
      default: () => {},
    },
    type: {
      type: String,
      default: '',
      require: true,
    },
    labelText: {
      type: String,
      default: '',
      require: true,
    },
    appType: {
      type: String,
      default: 'rag',
    },
  },
  data() {
    return {
      knowledgeList: [],
      knowledgeRecallConfig: {},
      currentKnowledgeId: '',
      knowledgeIndex: -1,
      currentMetaData: {},
      showGraphSwitch: false,
    };
  },
  watch: {
    type: {
      handler(val) {
        if (val === 'qaKnowledgeBaseConfig') {
          this.showGraphSwitch = false;
        }
      },
      immediate: true,
    },
    knowledgeConfig: {
      handler(val) {
        this.knowledgeList = val.knowledgebases || [];
        this.knowledgeRecallConfig = val.config || {};
        this.showGraphSwitch = this.knowledgeList.some(
          item => item.graphSwitch === 1,
        );
      },
      immediate: true,
      deep: true,
    },
  },
  computed: {
    showKnowledgeList() {
      return (
        this.knowledgeConfig &&
        this.knowledgeConfig.knowledgebases &&
        this.knowledgeConfig.knowledgebases.length
      );
    },
    // 是否全部选的是外部知识库
    isAllExternalKnowledgeSelected() {
      const knowledgebases = this.knowledgeConfig.knowledgebases || [];
      if (knowledgebases.length === 0) {
        return false;
      }
      return !knowledgebases.some(item => item.external !== 1);
    },
  },
  methods: {
    knowledgeRecallSet(data) {
      this.$emit('knowledgeRecallSet', data, this.type);
    },
    handleAdd() {
      this.$refs.knowledgeSelect.showDialog(this.knowledgeList);
    },
    getKnowledgeData(data) {
      this.$emit('getSelectKnowledge', data, this.type);
    },
    handleSetting(item, index) {
      this.currentKnowledgeId = item.id;
      this.knowledgeIndex = index;
      this.currentMetaData = {};
      this.$nextTick(() => {
        this.currentMetaData = item.metaDataFilterParams;
      });
      this.$refs.metaDataFilterField.showDialog();
    },
    handleDelete(index) {
      this.$emit('knowledgeDelete', index, this.type);
    },
    submitMetaData(data) {
      this.$emit('updateMetaData', data, this.knowledgeIndex, this.type);
    },
    showknowledgeRecallSet() {
      if (!this.knowledgeConfig.knowledgebases.length) return;
      this.$refs.knowledgeRecallField.showDialog();
    },
    // 跳转至知识库|问答库
    handleKnowledgeLink(category, id) {
      const targetRouterName =
        category === 1
          ? `/knowledge/qa/docList/${id}`
          : `/knowledge/doclist/${id}`;
      const routeData = this.$router.resolve({
        name: targetRouterName,
      });
      const fullUrl = `${routeData.href}${routeData.location.name}`.replace(
        /\/+/g,
        '/',
      );
      window.open(fullUrl, '_blank');
    },
  },
};
</script>

<style lang="scss" scoped>
.qa-database-container {
  width: 100%;
  .qa-database-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 5px;

    .header-left {
      display: flex;
      align-items: center;
      .required-label {
        width: 18px;
        height: 18px;
        margin-right: 4px;
      }
      .header-icon {
        font-size: 16px;
        color: #999;
      }

      .header-title {
        font-size: 15px;
        font-weight: bold;
      }
    }

    .header-right {
      display: flex;
      align-items: center;
      justify-content: flex-end;
      gap: 10px;
      .operation {
        cursor: pointer;
        font-size: 15px;
        .handleBtn {
          font-weight: bold;
        }
      }
      .common-add {
        display: flex;
        align-items: center;
        gap: 4px;
        cursor: pointer;
        font-size: 14px;
        font-weight: bold;
        .el-icon-plus {
          font-size: 14px;
        }

        .handleBtn {
          cursor: pointer;
        }

        &:hover {
          color: $color;
        }
      }
    }
  }

  .qa-database-content {
    .action-list {
      display: grid;
      gap: 10px;
      width: 100%;
      &.single-row {
        grid-template-columns: repeat(1, minmax(0, 1fr));
      }
      &.two-row {
        grid-template-columns: repeat(2, minmax(0, 1fr));
      }
      .action-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        background: #fafafa;
        border: 1px solid #e8e8e8;
        border-radius: 8px;
        padding: 10px 20px;
        box-sizing: border-box;

        .name {
          flex: 1;
          color: $color;
          font-size: 14px;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
          margin-right: 12px;
          cursor: pointer;
        }

        .bt {
          display: flex;
          align-items: center;
          justify-content: flex-end;
          flex-shrink: 0;

          .del {
            color: $color;
            font-size: 16px;
            cursor: pointer;
            transition: opacity 0.2s;

            &:hover {
              opacity: 0.7;
            }
          }
        }
      }
    }

    .empty-state {
      text-align: center;
      padding: 40px 0;
      color: #999;
      font-size: 14px;
    }
  }
}
</style>
