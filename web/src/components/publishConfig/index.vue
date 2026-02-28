<template>
  <div class="weburl-container page-wrapper right-page-content-body">
    <div class="weburl-title">
      <span class="el-icon-arrow-left goback" @click="goback"></span>
      <span class="weburl-title-text">
        {{ name }} - {{ $t('agent.form.publishConfig') }}
      </span>
    </div>
    <CommonLayout
      :showAside="true"
      :showTitle="false"
      :asideWidth="asideWidth"
      class="weburl-content"
    >
      <template #aside>
        <template v-for="(item, index) in toolList">
          <div
            v-if="
              (appType !== AGENT && item.type !== 'url') || appType === AGENT
            "
            :class="['toolList', item.type === active ? 'activeItem' : '']"
            @click="checkTool(item)"
            :key="'agent' + index"
          >
            <h3>{{ item.name }}</h3>
            <p>{{ item.desc }}</p>
          </div>
        </template>
      </template>
      <template #main-content>
        <CreateUrl
          ref="CreateUrl"
          v-if="active === 'url'"
          :appId="appId"
          :appType="appType"
        />
        <!--v0.3.3 发布配置不展示 API-->
        <!--<CreateApi
          ref="CreateApi"
          v-if="active === 'api'"
          :appId="appId"
          :appType="appType"
        />-->
        <CreateScope
          ref="CreateScope"
          v-if="active === 'scope'"
          :appId="appId"
          :appType="appType"
        />
      </template>
    </CommonLayout>
  </div>
</template>
<script>
import CommonLayout from '@/components/exploreContainer.vue';
import CreateApi from './createApi.vue';
import CreateUrl from './createUrl.vue';
import CreateScope from './createScope.vue';
import { AGENT } from '@/utils/commonSet';
export default {
  components: { CommonLayout, CreateApi, CreateUrl, CreateScope },
  data() {
    return {
      name: '',
      appId: '',
      appType: '',
      AGENT,
      active: 'url',
      asideWidth: '260px',
      toolList: [
        {
          name: 'Web URL',
          desc: this.$t('app.shareTip'),
          type: 'url',
        },
        /*{
          name: 'API',
          desc: '支持嵌入第三方应用系统',
          type: 'api',
        },*/
        {
          name: this.$t('list.version.publishType'),
          desc: this.$t('app.publishTypeDesc'),
          type: 'scope',
        },
      ],
    };
  },
  created() {
    const { appId, appType, name } = this.$route.query;
    this.appId = appId;
    this.appType = appType;
    this.name = name;
    this.active = appType === AGENT ? 'url' : 'scope';
  },
  methods: {
    checkTool(item) {
      this.active = item.type;
    },
    goback() {
      this.$router.back();
    },
  },
};
</script>
<style lang="scss" scoped>
.activeItem {
  border: 1px solid $color !important;
}
.weburl-container {
  width: 100%;
  height: 100%;
  padding: 0 10px;
  .weburl-title {
    width: 100%;
    height: 60px;
    padding: 0 20px;
    border-bottom: 1px solid #dbdbdb;
    display: flex;
    align-items: center;
    .goback {
      font-size: 20px;
      margin-right: 10px;
      cursor: pointer;
    }
    .weburl-title-text {
      font-size: 18px;
      font-weight: bold;
    }
  }
  .weburl-content {
    width: 100%;
    padding: 10px;
    height: calc(100% - 60px);
    .toolList {
      cursor: pointer;
      border: 1px solid #dbdbdb;
      text-align: center;
      width: 90%;
      height: 80px;
      border-radius: 6px;
      padding: 10px;
      margin: 20px auto;
      h3 {
        font-size: 16px;
      }
      p {
        color: #666;
        padding-top: 10px;
      }
    }
  }
  ::v-deep .explore-container {
    .page-wrapper {
      min-height: 0 !important;
      padding-left: 0;
    }
  }
}
</style>
