<template>
  <div>
    <div class="table-wrap list-common wrap-fullheight">
      <div class="table-box">
        <el-button
          class="add-bt"
          size="mini"
          type="primary"
          @click="exportData"
        >
          <span>{{ $t('common.button.export') }}</span>
        </el-button>
        <el-table
          :data="tableData"
          :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
          v-loading="loading"
          style="width: 100%"
        >
          <el-table-column
            prop="model"
            :label="$t('statisticsDashboard.modelName')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.model || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="provider"
            :label="$t('statisticsDashboard.provider')"
            align="left"
          >
            <template slot-scope="scope">
              {{ providerObj[scope.row.provider] || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="orgName"
            :label="$t('statisticsDashboard.org')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.orgName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="callCount"
            :label="
              $t('statisticsDashboard.callCount') +
              ` (${$t('statisticsDashboard.frequency')})`
            "
            align="left"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.callCount) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="callFailure"
            :label="
              $t('statisticsDashboard.callFailure') +
              ` (${$t('statisticsDashboard.frequency')})`
            "
            align="left"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.callFailure) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="failureRate"
            :label="$t('statisticsDashboard.failureRate')"
            align="left"
          />
          <el-table-column
            prop="promptTokens"
            :label="
              $t('statisticsDashboard.promptTokens') +
              ` (${$t('statisticsDashboard.quantity')})`
            "
            align="left"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.promptTokens) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="completionTokens"
            :label="
              $t('statisticsDashboard.completionTokens') +
              ` (${$t('statisticsDashboard.quantity')})`
            "
            align="left"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.completionTokens) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="totalTokens"
            :label="
              $t('statisticsDashboard.totalTokens') +
              ` (${$t('statisticsDashboard.quantity')})`
            "
            align="left"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.totalTokens) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="avgCosts"
            :label="$t('statisticsDashboard.avgCosts') + ` (ms)`"
            align="left"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.avgCosts) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="avgFirstTokenLatency"
            :label="$t('statisticsDashboard.avgFirstTokenLatency') + ` (ms)`"
            align="left"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.avgFirstTokenLatency) }}
            </template>
          </el-table-column>
        </el-table>
      </div>
      <Pagination
        class="pagination"
        ref="pagination"
        :listApi="listApi"
        @refreshData="refreshData"
      />
    </div>
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import { formatAmount, resDownloadFile } from '@/utils/util.js';
import { fetchModelList, exportModelData } from '@/api/statisticsDashboard';
import { PROVIDER_OBJ } from '@/views/modelAccess/constants';

export default {
  components: { Pagination },
  data() {
    return {
      listApi: fetchModelList,
      loading: false,
      tableData: [],
      providerObj: PROVIDER_OBJ,
    };
  },
  methods: {
    formatAmount,
    async getTableData(params) {
      if (this.$refs.pagination) {
        this.loading = true;
        try {
          this.tableData = await this.$refs.pagination.getTableData(params);
        } finally {
          this.loading = false;
        }
      }
    },
    refreshData(data) {
      this.tableData = data;
    },
    async exportData() {
      const response = await exportModelData(this.params);
      resDownloadFile(
        response,
        `${this.$t('statisticsDashboard.modelData')}.xlsx`,
      );
    },
  },
};
</script>

<style lang="scss" scoped>
.table-wrap {
  padding: 0 12px;
}
.table-box {
  .table-header {
    font-size: 16px;
    font-weight: bold;
    color: #555;
  }
  .add-bt {
    margin: 0 0 16px;
    float: right;
    img {
      width: 16px;
      margin-right: 5px;
      display: inline-block;
      vertical-align: middle;
    }
    span {
      display: inline-block;
      vertical-align: middle;
    }
  }
}
</style>
