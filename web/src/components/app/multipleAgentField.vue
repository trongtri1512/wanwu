<template>
  <div class="container">
    <div class="header">
      <div class="header-left">
        <span class="header-title">
          {{ labelText }}
        </span>
      </div>
      <div class="header-right">
        <span class="common-add" @click="handleAdd">
          <span class="el-icon-plus"></span>
          <span class="handleBtn">{{ $t('knowledgeSelect.add') }}</span>
        </span>
      </div>
    </div>
    <div class="content">
      <div class="list single-row" v-if="multiAgentList.length">
        <div v-for="(item, index) in multiAgentList" :key="index" class="item">
          <div
            style="
              display: flex;
              flex-direction: row;
              align-items: center;
              overflow: hidden;
            "
          >
            <div class="img">
              <img
                :src="avatarSrc(item.avatar.path)"
                v-if="item.avatar && item.avatar.path"
              />
            </div>
            <div class="info">
              <span class="name ellipsis">{{ item.name }}</span>
              <span class="desc ellipsis">{{ item.desc }}</span>
            </div>
          </div>
          <div class="bt">
            <el-tooltip
              effect="dark"
              :content="$t('agent.form.multiAgentConfig')"
              placement="top-start"
            >
              <span
                class="el-icon-setting del"
                @click="handleSetting(item)"
                style="margin-right: 10px"
              ></span>
            </el-tooltip>
            <el-switch
              v-model="item.enable"
              class="bt-switch"
              @change="handleSwitch(item)"
            ></el-switch>
            <span class="el-icon-delete del" @click="handleDelete(item)"></span>
          </div>
        </div>
      </div>
    </div>
    <multiAgentSelect
      ref="multiAgentSelect"
      :appId="appId"
      @bindAgent="bindAgent"
    />
    <multiAgentEditDialog ref="multiAgentEditDialog" @submit="editAgent" />
  </div>
</template>
<script>
import multiAgentSelect from '@/components/multiAgentSelect.vue';
import multiAgentEditDialog from '@/components/multiAgentEditDialog.vue';
import {
  unbindMultiAgent,
  switchMultiAgent,
  updateMultiAgent,
} from '@/api/agent';
import { avatarSrc } from '@/utils/util';

export default {
  components: { multiAgentSelect, multiAgentEditDialog },
  props: {
    multiAgentInfos: {
      type: Array,
      default: () => [],
      required: true,
    },
    labelText: {
      type: String,
      default: '',
      require: true,
    },
    appId: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      multiAgentList: [],
    };
  },
  watch: {
    multiAgentInfos: {
      handler(val) {
        this.multiAgentList = val;
      },
      immediate: true,
    },
  },
  methods: {
    avatarSrc,
    handleAdd() {
      this.$refs.multiAgentSelect.showDialog(this.multiAgentList);
    },
    bindAgent(item) {
      this.multiAgentList = [
        ...this.multiAgentList,
        {
          ...item,
          agentId: item.appId,
          enable: true,
        },
      ];
    },
    editAgent(item) {
      updateMultiAgent({
        agentId: item.agentId,
        assistantId: this.appId,
        desc: item.desc,
      })
        .then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            this.multiAgentList = this.multiAgentList.map(agent => {
              if (agent.agentId === item.agentId) {
                return item;
              }
              return agent;
            });
          }
        })
        .catch(() => {});
    },
    handleSetting(item) {
      this.$refs.multiAgentEditDialog.showDialog(item);
    },
    handleSwitch(item) {
      switchMultiAgent({
        agentId: item.agentId,
        assistantId: this.appId,
        enable: item.enable,
      })
        .then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
          }
        })
        .catch(() => {});
    },
    handleDelete(item) {
      unbindMultiAgent({
        agentId: item.agentId,
        assistantId: this.appId,
      })
        .then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            this.multiAgentList = this.multiAgentList.filter(
              agent => agent.agentId !== item.agentId,
            );
          }
        })
        .catch(() => {});
    },
  },
};
</script>

<style lang="scss" scoped>
.container {
  width: 100%;
  .header {
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

  .content {
    .list {
      display: grid;
      gap: 10px;
      width: 100%;
      &.single-row {
        grid-template-columns: repeat(1, minmax(0, 1fr));
      }
      &.two-row {
        grid-template-columns: repeat(2, minmax(0, 1fr));
      }
      .item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        background: #fafafa;
        border: 1px solid #e8e8e8;
        border-radius: 8px;
        padding: 10px 20px;
        box-sizing: border-box;

        .img {
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

        .info {
          flex: 1;
          color: #333;
          font-size: 14px;
          overflow: hidden;
          margin-right: 12px;
          display: flex;
          flex-direction: column;
          gap: 4px;

          .name {
            font-size: 14px;
            font-weight: 600;
            color: #1c1d23;
          }
          .desc {
            font-size: 12px;
            color: rgba(28, 29, 35, 0.8);
          }
        }

        .bt {
          display: flex;
          align-items: center;
          justify-content: flex-end;
          flex-shrink: 0;

          .bt-switch {
            margin-right: 10px;
          }

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
