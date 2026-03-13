<template>
  <div class="tempSquare-management">
    <div class="tempSquare-content-box tempSquare-third">
      <div class="tempSquare-main">
        <div class="tempSquare-content">
          <div class="tempSquare-card-box">
            <div class="card-search card-search-cust" v-if="!templateUrl">
              <div>
                <span
                  v-for="item in typeList"
                  :key="item.key"
                  :class="['tab-span', { 'is-active': typeRadio === item.key }]"
                  @click="changeTab(item.key)"
                >
                  {{ item.name }}
                </span>
              </div>
              <search-input
                style="margin-right: 2px"
                :placeholder="$t('tempSquare.searchText')"
                ref="searchInput"
                @handleSearch="doGetWorkflowTempList"
              />
            </div>

            <div class="card-loading-box" v-if="list.length && !templateUrl">
              <div class="card-box" v-loading="loading">
                <div
                  class="card"
                  v-for="(item, index) in list"
                  :key="index"
                  @click.stop="handleClick(item)"
                >
                  <div class="card-title">
                    <img
                      class="card-logo"
                      v-if="item.avatar && item.avatar.path"
                      :src="item.avatar.path"
                    />
                    <div class="mcp_detailBox">
                      <span class="mcp_name">{{ item.name }}</span>
                      <span class="mcp_from">
                        <label>
                          {{ $t('tempSquare.author') }}：{{ item.author }}
                        </label>
                      </span>
                    </div>
                  </div>
                  <div class="card-des">{{ item.desc }}</div>
                  <div class="card-bottom">
                    <div class="card-bottom-left">
                      {{ $t('tempSquare.downloadCount') }}：{{
                        item.downloadCount || 0
                      }}
                    </div>
                    <div class="card-bottom-right">
                      <el-tooltip
                        v-if="!isPublic"
                        :content="$t('tempSquare.copy')"
                        placement="top"
                      >
                        <i
                          class="el-icon-copy-document"
                          @click.stop="copyTemplate(item)"
                        ></i>
                      </el-tooltip>
                      <el-tooltip
                        :content="$t('tempSquare.download')"
                        placement="top"
                      >
                        <i
                          class="el-icon-download"
                          @click.stop="downloadTemplate(item)"
                        ></i>
                      </el-tooltip>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="empty">
              <el-empty :description="$t('common.noData')"></el-empty>
            </div>
          </div>
        </div>
      </div>
    </div>
    <HintDialog :templateUrl="templateUrl" ref="hintDialog" />
    <CreateWorkflow type="clone" ref="cloneWorkflowDialog" />
  </div>
</template>
<script>
import { getWorkflowTempList, downloadWorkflow } from '@/api/templateSquare';
import { resDownloadFile } from '@/utils/util';
import SearchInput from '@/components/searchInput.vue';
import HintDialog from './components/hintDialog.vue';
import CreateWorkflow from '@/components/createApp/createWorkflow.vue';
export default {
  components: { SearchInput, HintDialog, CreateWorkflow },
  props: {
    isPublic: true,
    type: '',
  },
  data() {
    return {
      basePath: this.$basePath,
      category: this.$t('square.all'),
      list: [],
      templateUrl: '',
      loading: false,
      typeRadio: 'all',
      typeList: [
        { name: this.$t('square.all'), key: 'all' },
        { name: this.$t('square.gov'), key: 'gov' },
        { name: this.$t('square.industry'), key: 'industry' },
        { name: this.$t('square.edu'), key: 'edu' },
        { name: this.$t('square.tourism'), key: 'tourism' },
        // {name: this.$t('square.medical'), key: 'medical'},
        { name: this.$t('square.data'), key: 'data' },
        { name: this.$t('square.creator'), key: 'create' },
        { name: this.$t('square.search'), key: 'search' },
      ],
    };
  },
  mounted() {
    this.doGetWorkflowTempList();
  },
  methods: {
    changeTab(key) {
      this.typeRadio = key;
      this.$refs.searchInput.value = '';
      this.doGetWorkflowTempList();
    },
    showHintDialog() {
      this.$refs.hintDialog.openDialog();
    },
    doGetWorkflowTempList() {
      const searchInput = this.$refs.searchInput;
      let params = {
        name: searchInput.value,
        category: this.typeRadio,
      };

      getWorkflowTempList(params)
        .then(res => {
          const { downloadLink = {}, list } = res.data || {};
          this.templateUrl = downloadLink.url;
          if (downloadLink.url) this.showHintDialog();

          this.list = list || [];
          this.loading = false;
        })
        .catch(() => (this.loading = false));
    },
    copyTemplate(item) {
      this.$refs.cloneWorkflowDialog.openDialog(item);
    },
    downloadTemplate(item) {
      downloadWorkflow({ templateId: item.templateId }).then(response => {
        resDownloadFile(response, `${item.name}.json`);
        this.doGetWorkflowTempList();
      });
    },
    handleClick(val) {
      const path = `${this.isPublic ? '/public' : ''}/templateSquare/detail`;
      this.$router.push({
        path,
        query: { templateSquareId: val.templateId, type: this.type },
      });
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tempSquare.scss';
</style>
