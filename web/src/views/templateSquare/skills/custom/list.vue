<template>
  <div class="tempSquare-management">
    <div class="tempSquare-content-box tempSquare-third">
      <div class="tempSquare-main">
        <div class="tempSquare-content">
          <div class="tempSquare-card-box">
            <div class="card-search card-search-cust">
              <search-input
                style="margin-right: 2px"
                :placeholder="$t('tempSquare.searchText')"
                ref="searchInput"
                @handleSearch="doGetSkillTempList"
              />
            </div>

            <div class="card-loading-box">
              <div class="card-box" v-loading="loading">
                <div class="card card-item-create" @click="handleAddSkill()">
                  <div class="app-card-create">
                    <div class="create-img-wrap">
                      <img
                        class="create-img"
                        src="@/assets/imgs/card_create_icon_skills.svg"
                        alt=""
                      />
                    </div>
                    <span>{{ $t('tempSquare.skills.app.addText') }}</span>
                  </div>
                </div>
                <skill-card
                  v-for="(item, index) in list"
                  :key="index"
                  :info="item"
                  :type="2"
                  @delete="handleDelete"
                  @download="handleDownload"
                />
              </div>
            </div>
            <div v-if="!list.length" class="empty">
              <el-empty :description="$t('common.noData')"></el-empty>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import SkillCard from '../card.vue';
import SearchInput from '@/components/searchInput.vue';
import { getCustomSkillList, deleteCustomSkill } from '@/api/templateSquare';
import { directDownload } from '@/utils/util';
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

      getCustomSkillList(params)
        .then(res => {
          const { list } = res.data || {};
          this.list = list || [];
          this.loading = false;
        })
        .catch(() => (this.loading = false));
    },
    handleDelete(info) {
      this.$confirm(
        this.$t('tempSquare.skills.deleteHint'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          dangerouslyUseHTMLString: true,
          type: 'warning',
          center: true,
        },
      ).then(async () => {
        deleteCustomSkill({
          skillId: info.skillId,
        }).then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.info.delete'));
            this.doGetSkillTempList();
          } else {
            this.$message.error(res.msg || this.$t('common.info.deleteErr'));
          }
        });
      });
    },
    handleAddSkill() {
      const path = '/skill/create';
      this.$router.push({
        path,
      });
    },
    handleDownload(info) {
      const { zipUrl } = info;
      directDownload(zipUrl);
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
  .card-bottom-right {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .card-item-create {
    min-height: 172px;
  }
}
</style>
