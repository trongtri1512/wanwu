<template>
  <div>
    <el-dialog
      :title="$t('agent.toolDialog.addTool')"
      :visible.sync="dialogVisible"
      width="50%"
      :before-close="handleClose"
    >
      <div class="tool-typ">
        <div class="toolbtn">
          <div
            v-for="(item, index) in toolList"
            :key="index"
            @click="clickTool(item, index)"
            :class="[{ active: activeValue === item.value }]"
          >
            {{ item.name }}
            <span>/</span>
            {{ showToolNum(item.value) }}
          </div>
        </div>
        <el-input
          v-model="toolName"
          :placeholder="$t('agent.toolDialog.searchTool')"
          class="tool-input"
          suffix-icon="el-icon-search"
          @keyup.enter.native="searchTool"
          clearable
        ></el-input>
      </div>
      <div class="toolContent">
        <div @click="goCreate" class="createTool">
          <span class="el-icon-plus add"></span>
          {{ createText() }}
        </div>
        <template v-for="(items, type) in contentMap">
          <template v-if="activeValue === type">
            <div
              v-for="item in items"
              :key="item[type + 'Id'] || item.id"
              class="toolContent_item"
            >
              <template v-if="type === 'workflow'">
                <div class="tool_box">
                  <div class="tool_img">
                    <img
                      :src="avatarSrc(item.avatar.path)"
                      v-if="item.avatar && item.avatar.path"
                    />
                  </div>
                  <div>
                    <div>{{ item.name }}</div>
                    <span class="tag" v-if="tagMap[item.appType]">
                      {{ tagMap[item.appType] }}
                    </span>
                  </div>
                </div>
                <div>
                  <el-button
                    type="text"
                    @click="openTool($event, item, type)"
                    v-if="!item.checked"
                  >
                    {{ $t('agent.toolDialog.add') }}
                  </el-button>
                  <el-button type="text" v-else style="color: #ccc">
                    {{ $t('agent.toolDialog.added') }}
                  </el-button>
                </div>
              </template>
              <el-collapse
                @change="handleToolChange"
                v-else
                class="tool_collapse"
              >
                <el-collapse-item :name="item.toolId">
                  <template slot="title">
                    <div class="tool_img">
                      <img
                        :src="avatarSrc(item.avatar.path)"
                        v-if="item.avatar && item.avatar.path"
                      />
                    </div>
                    <div :class="type === 'tool' && 'tool-name-container'">
                      <h3 class="tool-name">{{ item.toolName }}</h3>
                      <span
                        v-if="item.loading"
                        class="el-icon-loading loading-text"
                      ></span>
                      <span v-if="type === 'tool'" class="tag">
                        {{
                          item.toolType === 'builtin'
                            ? $t('agent.toolDialog.builtinTools')
                            : $t('agent.toolDialog.customTools')
                        }}
                      </span>
                    </div>
                  </template>
                  <template v-if="item.children && item.children.length">
                    <div
                      v-for="(tool, index) in item.children"
                      class="tool-action-item"
                      :key="'tool' + index"
                    >
                      <div style="padding-right: 5px">
                        <p>
                          <span>{{ tool.name }}</span>
                          <el-tooltip
                            class="item"
                            effect="dark"
                            :content="tool.description"
                            placement="top-start"
                            v-if="
                              tool.description && tool.description.length > 0
                            "
                          >
                            <span class="el-icon-info desc-info"></span>
                          </el-tooltip>
                        </p>
                      </div>
                      <div>
                        <el-button
                          type="text"
                          @click="openTool($event, item, type, tool)"
                          v-if="!tool.checked"
                        >
                          {{ $t('agent.toolDialog.add') }}
                        </el-button>
                        <el-button type="text" v-else style="color: #ccc">
                          {{ $t('agent.toolDialog.added') }}
                        </el-button>
                      </div>
                    </div>
                  </template>
                </el-collapse-item>
              </el-collapse>
            </div>
          </template>
        </template>
      </div>
    </el-dialog>
  </div>
</template>
<script>
import {
  addWorkFlowInfo,
  addMcp,
  addCustomBuiltIn,
  toolList,
  toolActionList,
  mcptoolList,
  mcpActionList,
  getWorkflowList,
} from '@/api/agent';
import { avatarSrc } from '@/utils/util';
export default {
  props: ['assistantId'],
  data() {
    return {
      toolName: '',
      dialogVisible: false,
      toolIndex: 0,
      activeValue: 'tool',
      workFlowInfos: [],
      mcpInfos: [],
      customInfos: [],
      mcpList: [],
      workFlowList: [],
      customList: [],
      builtInInfos: [],
      customCount: 0,
      mcpCount: 0,
      workflowCount: 0,
      toolList: [
        {
          value: 'tool',
          name: this.$t('agent.toolDialog.tool'),
        },
        {
          value: 'mcp',
          name: 'MCP',
        },
        {
          value: 'workflow',
          name: this.$t('appSpace.workflow'),
        },
      ],
    };
  },
  computed: {
    contentMap() {
      return {
        tool: this.customInfos,
        builtIn: this.builtInInfos,
        mcp: this.mcpInfos,
        workflow: this.workFlowInfos,
      };
    },
    tagMap() {
      return {
        workflow: this.$t('appSpace.workflow'),
        chatflow: this.$t('appSpace.chat'),
      };
    },
  },
  created() {
    this.getMcpSelect('');
    this.getWorkflowList('');
    this.getCustomList('');
  },
  methods: {
    avatarSrc,
    showToolNum(type) {
      if (type === 'tool') {
        return this.customCount;
      } else if (type === 'mcp') {
        return this.mcpCount;
      } else {
        return this.workflowCount;
      }
    },
    handleToolChange(id) {
      let toolId = id[0];
      if (this.activeValue === 'tool') {
        const targetItem = this.customInfos.find(
          item => item.toolId === toolId,
        );
        if (targetItem) {
          const { toolId, toolType } = targetItem;
          const index = this.customInfos.findIndex(
            item => item.toolId === toolId,
          );
          this.getToolAction(toolId, toolType, index);
        }
      } else if (this.activeValue === 'mcp') {
        const targetItem = this.mcpInfos.find(item => item.toolId === toolId);
        if (targetItem) {
          const { toolId, toolType } = targetItem;
          const index = this.mcpInfos.findIndex(item => item.toolId === toolId);
          this.getMcpAction(toolId, toolType, index);
        }
      }
    },
    getCustomList(name) {
      //获取自定义和内置工具
      toolList({ name }).then(res => {
        if (res.code === 0) {
          this.customInfos = (res.data.list || []).map(m => ({
            ...m,
            loading: false,
            children: [],
          }));
        }
      });
    },
    getToolAction(toolId, toolType, index) {
      this.$set(this.customInfos[index], 'loading', true);
      toolActionList({ toolId, toolType })
        .then(res => {
          if (res.code === 0) {
            this.$set(this.customInfos[index], 'children', res.data.actions);
            this.$set(this.customInfos[index], 'loading', false);
            this.customInfos[index]['children'].forEach(m => {
              m.checked = this.customList.some(
                item => item.actionName === m.name && item.toolId === toolId,
              );
            });
          }
        })
        .catch(() => {
          this.$set(this.customInfos[index], 'loading', false);
        });
    },
    goCreate() {
      if (this.activeValue === 'tool') {
        this.$router.push({ path: '/tool?tool=custom' });
      } else if (this.activeValue === 'mcp') {
        this.$router.push({ path: '/mcpService?mcp=integrate' });
      } else {
        this.$router.push({ path: '/appSpace/workflow' });
      }
    },
    createText() {
      if (this.activeValue === 'tool') {
        return this.$t('agent.toolDialog.createAutoTool');
      } else if (this.activeValue === 'mcp') {
        return this.$t('agent.toolDialog.importMcp');
      } else {
        return this.$t('agent.toolDialog.createWorkflow');
      }
    },
    openTool(e, item, type, action) {
      if (!e) return;
      if (type === 'workflow') {
        this.addWorkFlow(item);
      } else if (type === 'mcp') {
        this.addMcpItem(item, action);
      } else {
        if (item.needApiKeyInput && !item.apiKey.length) {
          this.$message.warning(this.$t('agent.toolDialog.errorApiKey'));
        }
        this.addCustomBuiltIn(item, action);
      }
    },
    addCustomBuiltIn(n, action) {
      //添加自定义工具和内置工具
      addCustomBuiltIn({
        assistantId: this.assistantId,
        actionName: action.name,
        toolId: n.toolId,
        toolType: n.toolType,
      }).then(res => {
        if (res.code === 0) {
          this.$set(action, 'checked', true);
          this.customCount++;
          this.$forceUpdate();
          this.$message.success(this.$t('agent.toolDialog.addSuccess'));
          this.$emit('updateDetail');
        }
      });
    },
    addMcpItem(n, action) {
      addMcp({
        assistantId: this.assistantId,
        actionName: action.name,
        mcpId: n.toolId,
        mcpType: n.toolType,
      }).then(res => {
        if (res.code === 0) {
          this.$set(action, 'checked', true);
          this.mcpCount++;
          this.$forceUpdate();
          this.$message.success(this.$t('agent.toolDialog.addSuccess'));
          this.$emit('updateDetail');
        }
      });
    },
    addWorkFlow(n) {
      this.doCreateWorkFlow(n, n.appId);
    },
    async doCreateWorkFlow(n, workFlowId, schema) {
      let params = {
        assistantId: this.assistantId,
        workFlowId,
      };
      let res = await addWorkFlowInfo(params);
      if (res.code === 0) {
        n.checked = true;
        this.workflowCount++;
        this.$message.success(this.$t('agent.addWorkFlowTips'));
        this.$emit('updateDetail');
      }
    },
    searchTool() {
      if (this.activeValue === 'tool') {
        this.getCustomList(this.toolName);
      } else if (this.activeValue === 'mcp') {
        this.getMcpSelect(this.toolName);
      } else {
        this.getWorkflowList(this.toolName);
      }
    },
    getMcpSelect(name) {
      //获取mcp工具
      mcptoolList({ name }).then(res => {
        if (res.code === 0) {
          this.mcpInfos = (res.data.list || []).map(m => ({
            ...m,
            children: [],
            loading: false,
          }));
        }
      });
    },
    getMcpAction(toolId, toolType, index) {
      this.$set(this.mcpInfos[index], 'loading', true);
      mcpActionList({ toolId, toolType })
        .then(res => {
          if (res.code === 0) {
            this.$set(this.mcpInfos[index], 'children', res.data.actions);
            this.$set(this.mcpInfos[index], 'loading', false);
            this.mcpInfos[index]['children'].forEach(m => {
              m.checked = this.mcpList.some(
                item => item.actionName === m.name && item.mcpId === toolId,
              );
            });
          }
        })
        .catch(() => {
          this.$set(this.mcpInfos[index], 'loading', false);
        });
    },
    getWorkflowList(name) {
      getWorkflowList({ name }).then(res => {
        if (res.code === 0) {
          this.workFlowInfos = (res.data.list || []).map(m => ({
            ...m,
            checked: this.workFlowList.some(
              item => item.workFlowId === m.appId,
            ),
          }));
        }
      });
    },
    showDialog(row) {
      this.dialogVisible = true;
      this.setWorkflow(row.workFlowInfos);
      this.mcpList = row.mcpInfos || [];
      this.workFlowList = row.workFlowInfos || [];
      this.customList = row.customInfos || [];
      this.customCount = this.customList.length;
      this.mcpCount = this.mcpList.length;
      this.workflowCount = this.workFlowList.length;
    },
    setWorkflow(data) {
      this.workFlowInfos = this.workFlowInfos.map(m => ({
        ...m,
        checked: data.some(item => item.workFlowId === m.appId),
      }));
    },
    handleClose() {
      this.toolIndex = -1;
      this.activeValue = 'tool';
      this.dialogVisible = false;
    },
    clickTool(item, i) {
      this.toolIndex = i;
      this.activeValue = item.value;
      if (this.activeValue === 'tool') {
        this.getCustomList('');
      } else if (this.activeValue === 'mcp') {
        this.getMcpSelect('');
      } else {
        this.getWorkflowList('');
      }
    },
  },
};
</script>
<style lang="scss" scoped>
::v-deep {
  .el-dialog__body {
    padding: 10px 20px;
  }
  .tool_collapse {
    width: 100% !important;
    border: none !important;
  }
  .el-collapse-item__header {
    background: none !important;
    border-bottom: none !important;
  }
  .el-collapse-item__wrap {
    border-bottom: none !important;
    background: none !important;
  }
  .el-collapse-item__content {
    padding-bottom: 0 !important;
  }
}
.createTool {
  padding: 10px;
  cursor: pointer;
  .add {
    padding-right: 5px;
  }
}
.createTool:hover {
  color: $color;
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
  max-height: 400px;
  overflow-y: auto;
  .toolContent_item {
    padding: 5px 20px;
    border: 1px solid #dbdbdb;
    border-radius: 6px;
    margin-bottom: 10px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: space-between;

    .tag {
      padding: 0 8px;
      background: rgba(139, 139, 149, 0.15);
      color: #4b4a58;
      font-size: 12px;
      border-radius: 6px;
    }
    .tool_box {
      display: flex;
      align-items: center;
    }
    .tool_img {
      width: 35px;
      height: 35px;
      background: #eee;
      border-radius: 50%;
      display: inline-block;
      margin-right: 5px;
      img {
        width: 100%;
        height: 100%;
        border-radius: 50%;
        object-fit: cover;
      }
    }
    .loading-text {
      margin-left: 4px;
      color: var($color);
    }
    .tool-action-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-top: 1px solid #eee;
      padding: 5px 0;
      .desc-info {
        color: #ccc;
        margin-left: 4px;
        cursor: pointer;
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

.tool-name-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
  .tool-name {
    height: auto;
    line-height: 1.5;
  }
  .tag {
    height: auto;
    line-height: 1.5;
    width: fit-content;
  }
}
</style>
