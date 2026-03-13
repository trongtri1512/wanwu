<template>
  <div
    class="tempSquare-detail page-wrapper"
    :style="isPublic ? `background: ${bgColor}; min-height: 100%` : ''"
  >
    <span class="back" @click="back">
      {{
        $t('menu.back') +
        (type === workflow ? $t('menu.templateSquare') : $t('menu.resource'))
      }}
    </span>
    <div class="tempSquare-title">
      <div class="tempSquare-title-left">
        <img
          class="logo"
          v-if="detail.avatar && detail.avatar.path"
          :src="
            type === workflow
              ? detail.avatar.path
              : avatarSrc(detail.avatar.path)
          "
        />
        <div :class="['info', { fold: foldStatus }]">
          <p class="name">{{ detail.name }}</p>
          <p v-if="detail.desc && detail.desc.length > 260" class="desc">
            {{ foldStatus ? detail.desc : detail.desc.slice(0, 268) + '...' }}
            <span class="arrow" v-show="detail.desc.length > 260" @click="fold">
              {{
                foldStatus
                  ? $t('common.button.fold')
                  : $t('common.button.detail')
              }}
            </span>
          </p>
          <p v-else class="desc">{{ detail.desc }}</p>
        </div>
      </div>
      <div style="margin-left: 10px">
        <el-button
          v-if="type === workflow"
          type="primary"
          size="mini"
          @click="copyTemplate(detail)"
        >
          {{ $t('tempSquare.copy') }}
        </el-button>
        <el-button type="primary" size="mini" @click="downloadTemplate(detail)">
          {{ $t('tempSquare.download') }}
        </el-button>
      </div>
    </div>
    <div class="tempSquare-main">
      <div class="left-info">
        <div class="tempSquare-tabs">
          <div
            :class="['tempSquare-tab', { active: tabActive === 0 }]"
            @click="tabClick(0)"
          >
            {{ $t('square.info') }}
          </div>
        </div>

        <div>
          <div
            class="overview bg-border"
            v-if="detail.summary || detail.feature || detail.scenario"
          >
            <div class="overview-item" v-if="detail.summary">
              <div class="item-title">• &nbsp;{{ $t('square.summary') }}</div>
              <div class="item-desc" v-html="parseTxt(detail.summary)"></div>
            </div>
            <div class="overview-item" v-if="detail.feature">
              <div class="item-title">• &nbsp;{{ $t('square.feature') }}</div>
              <div class="item-desc" v-html="parseTxt(detail.feature)"></div>
            </div>
            <div class="overview-item" v-if="detail.scenario">
              <div class="item-title">• &nbsp;{{ $t('square.scenario') }}</div>
              <div class="item-desc">
                <div v-html="parseTxt(detail.scenario)"></div>
              </div>
            </div>
          </div>
          <div class="overview bg-border" v-if="detail.note">
            <div class="overview-item">
              <div class="item-title">• &nbsp;{{ $t('square.note') }}</div>
              <div class="item-desc" v-html="parseTxt(detail.note)"></div>
            </div>
          </div>
          <div class="overview bg-border" v-if="detail.skillMarkdown">
            <div class="overview-item">
              <!--<div class="item-title">• &nbsp;{{ $t('square.detail') }}</div>-->
              <div class="item-desc">
                <div class="tempSquare-markdown">
                  <MdRender :content="detail.skillMarkdown" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="right-recommend">
        <p style="margin: 20px 0; color: #333">
          {{ $t('tempSquare.otherTemp') }}
        </p>
        <div
          class="recommend-item"
          v-for="(item, i) in recommendList"
          :key="`${i}rc`"
          @click="handleClick(item)"
        >
          <img
            class="logo"
            v-if="item.avatar && item.avatar.path"
            :src="
              type === workflow ? item.avatar.path : avatarSrc(item.avatar.path)
            "
          />
          <p class="name">{{ item.name }}</p>
          <p class="intro">{{ item.desc }}</p>
        </div>
      </div>
    </div>
    <CreateWorkflow type="clone" ref="cloneWorkflowDialog" />
  </div>
</template>
<script>
import {
  downloadWorkflow,
  getWorkflowRecommendsList,
  getWorkflowTempInfo,
  getSkillTempInfo,
  getSkillTempList,
  downloadSkill,
  getCustomSkillInfo,
  getCustomSkillList,
  downloadCustomSkill,
} from '@/api/templateSquare';
import { SKILL, WORKFLOW, SKILLCUSTOM } from './constants';
import { avatarSrc, directDownload, resDownloadFile } from '@/utils/util';
import CreateWorkflow from '@/components/createApp/createWorkflow.vue';
import MdRender from '@/components/mdRender.vue';

export default {
  components: { CreateWorkflow, MdRender },
  data() {
    return {
      basePath: this.$basePath,
      isPublic: true,
      bgColor:
        'linear-gradient(1deg, rgb(247, 252, 255) 50%, rgb(233, 246, 254) 98%)',
      type: '',
      workflow: WORKFLOW,
      isFromSquare: true,
      templateSquareId: '',
      detail: {},
      foldStatus: false,
      tabActive: 0,
      recommendList: [],
      dialogVisible: false,
    };
  },
  watch: {
    $route: {
      handler() {
        this.initData();
      },
      // 深度观察监听
      deep: true,
    },
  },
  created() {
    this.isPublic = this.$route.path.includes('/public/');
  },
  mounted() {
    this.initData();
    this.getRecommendList();
  },
  methods: {
    avatarSrc,
    initData() {
      const { type, templateSquareId } = this.$route.query || {};
      this.templateSquareId = templateSquareId;
      this.type = type || WORKFLOW;
      this.getDetailData();

      // 滚动到顶部
      const main = document.querySelector('.el-main > .page-container');
      if (main) main.scrollTop = 0;
    },
    async getDetailData() {
      let res;
      if (this.type === WORKFLOW) {
        res = await getWorkflowTempInfo({ templateId: this.templateSquareId });
      } else if (this.type === SKILLCUSTOM) {
        res = await getCustomSkillInfo({ skillId: this.templateSquareId });
      } else {
        res = await getSkillTempInfo({ skillId: this.templateSquareId });
      }
      this.detail = res.data || {};
    },
    async getRecommendList() {
      let res;
      if (this.type === WORKFLOW) {
        res = await getWorkflowRecommendsList({
          templateId: this.templateSquareId,
        });
      } else if (this.type === SKILLCUSTOM) {
        res = await getCustomSkillList();
      } else {
        res = await getSkillTempList();
      }
      this.recommendList = res.data.list || [];
    },
    copyTemplate(item) {
      this.$refs.cloneWorkflowDialog.openDialog(item);
    },
    async downloadTemplate(item) {
      const isWorkflow = this.type === WORKFLOW;
      let res;
      if (isWorkflow) {
        res = await downloadWorkflow({ templateId: item.templateId });
      } else if (this.type === SKILLCUSTOM) {
        await this.handleDownloadCustomSkill(item);
        return;
      } else {
        res = await downloadSkill({ skillId: item.skillId });
      }
      resDownloadFile(res, `${item.name}${isWorkflow ? '.json' : '.zip'}`);
    },
    getPath() {
      return this.type === SKILL || this.type === SKILLCUSTOM
        ? '/skill'
        : this.isPublic
          ? '/public/templateSquare'
          : '/templateSquare';
    },
    handleClick(val) {
      const templateSquareId =
        this.type === WORKFLOW ? val.templateId : val.skillId;
      this.$router.push(
        `${this.getPath()}/detail?templateSquareId=${templateSquareId}&type=${this.type}`,
      );
    },
    // 解析文本，遇到.换行等
    parseTxt(txt) {
      if (!txt) return '';
      const text = txt
        .replaceAll('\n\t', '<br/>&nbsp;')
        .replaceAll('\n', '<br/>')
        .replaceAll('\t', '   &nbsp;');
      return text;
    },
    tabClick(status) {
      this.tabActive = status;
    },
    fold() {
      this.foldStatus = !this.foldStatus;
    },
    back() {
      this.$router.push({ path: this.getPath(), query: { type: this.type } });
    },
    // 自定义skills下载
    handleDownloadCustomSkill(skillInfo) {
      const { zipUrl } = skillInfo;
      directDownload(zipUrl);
    },
  },
};
</script>
<style lang="scss">
.tempSquare-detail {
  padding: 20px;
  overflow: auto;
  .back {
    color: $color;
    cursor: pointer;
  }
  .tempSquare-title {
    padding: 20px 0;
    display: flex;
    border-bottom: 1px solid #bfbfbf;
    justify-content: space-between;
    align-items: center;
    .tempSquare-title-left {
      display: flex;
      align-items: center;
    }
    .logo {
      width: 54px;
      height: 54px;
      object-fit: cover;
    }
    .info {
      position: relative;
      margin-left: 15px;
      .name {
        font-size: 16px;
        color: #5d5d5d;
        font-weight: bold;
      }
      .desc {
        margin-top: 10px;
        line-height: 22px;
        color: #9f9f9f;
        word-break: break-all;
      }
      .arrow {
        position: absolute;
        display: block;
        right: 0;
        bottom: -5px;
        cursor: pointer;
        color: $color;
        margin-left: 10px;
        font-size: 13px;
      }
    }
    .fold {
      height: auto;
    }
  }
  .tempSquare-main {
    display: flex;
    margin: 10px 0 0 0;
    .left-info {
      width: calc(100% - 420px);
      margin-right: 20px;
      .tempSquare-tabs {
        margin: 20px 0 0 0;
        .tempSquare-tab {
          display: inline-block;
          vertical-align: middle;
          width: 160px;
          height: 40px;
          border-bottom: 1px solid #333;
          line-height: 40px;
          text-align: center;
          cursor: pointer;
        }
        .active {
          background: #333;
          color: #fff;
          font-weight: bold;
        }
      }
      .overview {
        .overview-item {
          display: flex;
          padding: 15px 0;
          border-bottom: 1px solid #eee;
          line-height: 24px;
          .item-title {
            width: 80px;
            color: $color;
            font-weight: bold;
          }
          .item-desc {
            width: calc(100% - 100px);
            margin-left: 10px;
            flex: 1;
            color: #333;
          }
        }
        .overview-item:last-child {
          border-bottom: none;
        }
      }
    }
    .right-recommend {
      width: 400px;
      overflow-y: auto;
      border-left: 1px solid #eee;
      padding: 20px;
      max-height: 900px;
      .recommend-item {
        position: relative;
        border: 1px solid $border_color;
        background: $color_opacity;
        margin-bottom: 15px;
        border-radius: 10px;
        padding: 20px 20px 20px 80px;
        text-align: left;
        cursor: pointer;
        min-height: 100px;
        .logo {
          width: 46px;
          height: 46px;
          object-fit: cover;
          position: absolute;
          left: 20px;
          border: 1px solid #fff;
          border-radius: 4px;
        }
        .name {
          color: #5d5d5d;
          font-weight: bold;
        }
        .intro {
          max-height: 36px;
          color: #5d5d5d;
          margin-top: 8px;
          font-size: 13px;
          overflow: hidden;
          display: -webkit-box;
          -webkit-box-orient: vertical;
          text-overflow: ellipsis;
          -webkit-line-clamp: 2;
          line-clamp: 2;
        }
      }
    }
  }
  .bg-border {
    margin-top: 20px;
    background-color: rgba(255, 255, 255, 1);
    box-sizing: border-box;
    border-radius: 10px;
    padding: 10px 20px;
    box-shadow: 2px 2px 15px $color_opacity;
  }
  .overview-item .item-desc {
    line-height: 28px;
  }
}
.tempSquare-markdown {
  ::v-deep.code-header {
    padding: 0 0 5px 0;
  }
}
</style>
