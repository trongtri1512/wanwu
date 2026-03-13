<template>
  <div class="tempSquare-management">
    <div class="tempSquare-content-box tempSquare-third">
      <div class="tempSquare-main">
        <div class="tempSquare-content">
          <div class="tempSquare-card-box">
            <div class="card-search card-search-cust">
              <SearchInput
                style="margin-right: 2px"
                :placeholder="$t('tempSquare.searchText')"
                ref="searchInput"
                @handleSearch="doGetSkillTempList"
              />
            </div>

            <div class="card-loading-box" v-if="list.length">
              <div class="card-box" v-loading="loading">
                <skill-card
                  v-for="(item, index) in list"
                  :key="index"
                  :info="item"
                  :type="1"
                  @download="handleDownload"
                />
                <div class="card card-item-more" @click="handleLinkMore()">
                  <div class="card-content">
                    <span>{{ $t('tempSquare.skills.app.moreText') }}</span>
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
  </div>
</template>
<script>
import SkillCard from './card.vue';
import { resDownloadFile } from '@/utils/util';
import SearchInput from '@/components/searchInput.vue';
import { getSkillTempList, downloadSkill } from '@/api/templateSquare';

export default {
  components: { SearchInput, SkillCard },
  props: {
    type: '',
  },
  data() {
    return {
      basePath: this.$basePath,
      list: [],
      templateUrl: '',
      loading: false,
    };
  },
  mounted() {
    this.doGetSkillTempList();
  },
  methods: {
    doGetSkillTempList() {
      const searchInput = this.$refs.searchInput;
      const params = {
        name: searchInput.value,
      };

      getSkillTempList(params)
        .then(res => {
          const { list } = res.data || {};
          this.list = list || [];
          this.loading = false;
        })
        .catch(() => (this.loading = false));
    },
    handleDownload(info) {
      downloadSkill({ skillId: info.skillId }).then(response => {
        resDownloadFile(response, `${info.name}.zip`);
        this.doGetSkillTempList();
      });
    },
    handleLinkMore() {
      window.open('https://clawhub.ai/skills?sort=downloads', '_blank');
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tempSquare.scss';
.tempSquare-management {
  .card-search-cust {
    justify-content: flex-start;
    margin-top: 10px;
  }

  .card-item-more {
    display: flex;
    height: auto !important;
    justify-content: center;
    align-items: center;
    min-height: 140px;
    .card-content {
      font-size: 16px;
      font-weight: 500;
      color: #5d5d5d;
      &:hover {
        color: $color;
      }
    }
  }
}
</style>
