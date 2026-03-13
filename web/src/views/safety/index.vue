<template>
  <div class="page-wrapper">
    <!--<div class="page-title">
      <img class="page-title-img" src="@/assets/imgs/safety.svg" alt="" />
      <span class="page-title-name">{{ $t('safety.title') }}</span>
    </div>-->
    <div style="padding: 20px">
      <div class="tabs" style="margin: 0">
        <div
          v-for="item in !isSystem
            ? tabList.filter(({ type }) => type === 'personal')
            : tabList"
          :key="item.type"
          :class="['tab', { active: type === item.type }]"
          @click="tabClick(item.type)"
        >
          {{ item.name }}
        </div>
        <p class="page-tips">{{ $t('safety.tips') }}</p>
      </div>
      <safetyList
        :appData="safetyData"
        @editItem="showCreate"
        @reloadData="getTableData"
        ref="knowledgeList"
        v-loading="tableLoading"
      />
      <createSafety
        ref="createSafety"
        @reloadData="getTableData"
        :type="type"
      />
    </div>
  </div>
</template>
<script>
import { getSensitiveList } from '@/api/safety';
import safetyList from './component/safetyList.vue';
import createSafety from './component/create.vue';
export default {
  components: { safetyList, createSafety },
  data() {
    return {
      isSystem: this.$store.state.user.permission.isSystem || false,
      type: 'personal',
      safetyData: [],
      tableLoading: false,
      tabList: [
        { name: this.$t('safety.personal'), type: 'personal' },
        { name: this.$t('safety.global'), type: 'global' },
      ],
    };
  },
  mounted() {
    const { type } = this.$route.query || {};
    this.type = type || 'personal';
    this.getTableData();
  },
  methods: {
    tabClick(type) {
      this.type = type;
      this.getTableData();
      if (type === 'personal') {
        this.$router.replace({ query: {} });
      } else {
        this.$router.replace({ query: { type } });
      }
    },
    getTableData() {
      this.tableLoading = true;
      getSensitiveList({ type: this.type })
        .then(res => {
          this.safetyData = res.data.list || [];
          this.tableLoading = false;
        })
        .catch(error => {
          this.tableLoading = false;
          this.$message.error(error);
        });
    },
    showCreate(row) {
      this.$refs.createSafety.showDialog(row);
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/tabs.scss';
.search-box {
  display: flex;
  justify-content: space-between;
}

::v-deep {
  .el-loading-mask {
    background: none !important;
  }
}
.page-tips {
  color: #888888;
  padding-top: 15px;
  font-weight: normal;
}
</style>
