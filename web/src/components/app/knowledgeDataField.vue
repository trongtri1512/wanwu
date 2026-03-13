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
            <img
              class="avatar"
              :src="
                avatarSrc(
                  item.avatar.path,
                  require('@/assets/imgs/knowledgeIcon.png'),
                )
              "
            />
            <div>
              <span>
                {{ item.name }}
              </span>
              <div class="knowledge-meta">
                <span class="meta-text">
                  {{
                    item.share
                      ? $t('knowledgeManage.public')
                      : $t('knowledgeManage.private')
                  }}
                </span>
                <span v-if="item.share" class="meta-text">
                  {{ item.orgName }}
                </span>
                <span v-if="item.external === 1" class="meta-text">
                  {{ $t('knowledgeManage.ribbon.external') }}
                </span>
                <span v-if="item.category === 2" class="meta-text">
                  {{ $t('knowledgeManage.ribbon.multimodal') }}
                </span>
              </div>
            </div>
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
import { avatarSrc } from '@/utils/util';
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
        console.log('knowledgeList', this.knowledgeList);
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
    avatarSrc,
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
      const baseUrl = routeData.href;
      const locationName = routeData.location.name;
      const combinedUrl = baseUrl + locationName;
      const fullUrl = combinedUrl.replace(/\/+/g, '/');
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
          display: flex;
          align-items: center;
          gap: 12px;
          color: #333;
          font-size: 14px;
          cursor: pointer;

          &:hover {
            color: $color;
          }

          img {
            width: 40px;
            height: 40px;
            border-radius: 6px;
            background: #f0f0f0;
            object-fit: cover;
          }

          .knowledge-meta {
            display: flex;
            gap: 8px;
            margin-top: 5px;
            span {
              padding: 2px 8px;
              background: rgba(139, 139, 149, 0.15);
              color: #4b4a58;
              font-size: 12px;
              border-radius: 6px;
            }
          }
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
