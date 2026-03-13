<template>
  <div>
    <el-dialog
      :visible.sync="dialogVisible"
      width="40%"
      :before-close="handleClose"
    >
      <template slot="title">
        <div class="dialog_title">
          <h3>
            {{
              category === 0
                ? $t('knowledgeSelect.title')
                : $t('app.selectQAdatabase')
            }}
          </h3>
          <el-input
            v-model="toolName"
            :placeholder="
              category === 0
                ? $t('knowledgeSelect.searchPlaceholder')
                : $t('app.qaSearchPlaceholder')
            "
            class="tool-input"
            suffix-icon="el-icon-search"
            clearable
            @keyup.enter.native="searchTool"
            @clear="searchTool"
          ></el-input>
        </div>
      </template>
      <div class="toolContent">
        <div
          v-for="item in knowledgeData"
          :key="item['knowledgeId']"
          class="toolContent_item"
        >
          <div style="display: flex; flex-direction: column; gap: 4px">
            <div style="display: flex; align-items: center; gap: 10px">
              <img
                class="avatar"
                :src="
                  avatarSrc(
                    item.avatar?.path,
                    require('@/assets/imgs/knowledgeIcon.png'),
                  )
                "
              />
              <div class="knowledge-info">
                <span class="knowledge-name">{{ item.name }}</span>
                <span class="knowledge-desc">{{ item.description }}</span>
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
            <span class="knowledge-createAt">
              {{ $t('knowledgeSelect.createTime') }} {{ item.createAt }}
            </span>
          </div>
          <el-button
            type="primary"
            @click="openTool($event, item)"
            v-if="!item.checked"
            size="small"
          >
            {{ $t('knowledgeSelect.add') }}
          </el-button>
          <el-button type="primary" v-else size="small">
            {{ $t('knowledgeSelect.added') }}
          </el-button>
        </div>
      </div>
      <span slot="footer" class="dialog-footer">
        <el-button @click="handleClose">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button type="primary" @click="submit">
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>
<script>
import { getKnowledgeList } from '@/api/knowledge';
import { avatarSrc } from '@/utils/util';
export default {
  props: ['category'],
  data() {
    return {
      dialogVisible: false,
      knowledgeData: [],
      knowledgeList: [],
      checkedData: [],
      toolName: '',
      selectedItems: [],
    };
  },
  created() {
    this.getKnowledgeList('');
  },
  methods: {
    avatarSrc,
    getKnowledgeList(name) {
      getKnowledgeList({ name, category: this.category, external: -1 })
        .then(res => {
          if (res.code === 0) {
            this.knowledgeData = (res.data.knowledgeList || []).map(m => ({
              ...m,
              checked: this.selectedItems.some(
                item => item.id === m.knowledgeId,
              ),
            }));
          }
        })
        .catch(() => {});
    },
    openTool(e, item) {
      if (!e) return;
      item.checked = !item.checked;
      if (item.checked) {
        const exists = this.selectedItems.find(
          si => si.id === item.knowledgeId,
        );
        if (!exists) {
          this.selectedItems.push({
            id: item.knowledgeId,
            name: item.name,
            graphSwitch: item.graphSwitch,
            external: item.external,
            category: item.category,
          });
        }
      } else {
        this.selectedItems = this.selectedItems.filter(
          si => si.id !== item.knowledgeId,
        );
      }
    },
    searchTool() {
      this.getKnowledgeList(this.toolName);
    },
    showDialog(data) {
      this.dialogVisible = true;
      this.selectedItems = JSON.parse(JSON.stringify(data || []));
      this.searchTool();
    },
    handleClose() {
      this.toolName = '';
      this.dialogVisible = false;
    },
    submit() {
      this.$emit('getKnowledgeData', this.selectedItems);
      this.toolName = '';
      this.dialogVisible = false;
    },
  },
};
</script>
<style lang="scss" scoped>
::v-deep {
  .el-dialog__body {
    padding: 10px 20px;
  }
  .el-dialog__header {
    display: flex;
    align-items: center;
    .el-dialog__headerbtn {
      top: unset !important;
    }
  }
}
.dialog_title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex: 1;
  h3 {
    font-size: 16px;
    font-weight: bold;
  }
  .tool-input {
    width: 250px;
    margin-right: 28px;
  }
}
.tool-typ {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid #dbdbdb;
  .toolbtn {
    display: flex;
    justify-content: flex-start;
    gap: 20px;
    div {
      text-align: center;
      padding: 5px 20px;
      border-radius: 6px;
      border: 1px solid #ddd;
      cursor: pointer;
    }
  }
  .tool-input {
    width: 200px;
  }
}
.toolContent {
  padding: 10px 0;
  max-height: 300px;
  overflow-y: auto;
  .toolContent_item {
    padding: 5px 20px;
    border-bottom: 1px solid $color_opacity;
    border-radius: 6px;
    margin-bottom: 10px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: space-between;
    img {
      width: 40px;
      height: 40px;
      border-radius: 6px;
      background: #f0f0f0;
      object-fit: cover;
    }
    ::v-deep {
      .el-button--primary {
        background: #fff !important;
        border: 1px solid #eee !important;
        padding: 8px 16px;
        border-radius: 6px;
        span {
          color: $color !important;
          font-size: 14px;
        }
      }
    }
    .knowledge-info {
      display: flex;
      flex-direction: column;
      gap: 4px;
      .knowledge-name {
        font-size: 14px;
        font-weight: 600;
        color: #1c1d23;
      }
      .knowledge-desc,
      .knowledge-createAt {
        font-size: 12px;
      }
      .knowledge-desc {
        color: rgba(28, 29, 35, 0.8);
      }
      .knowledge-createAt {
        color: rgba(28, 29, 35, 0.35);
        margin-top: 5px;
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
  }
  .toolContent_item:hover {
    background: $color_opacity;
  }
}
.active {
  border: 1px solid $color !important;
  color: #fff;
  background: $color;
}
</style>
