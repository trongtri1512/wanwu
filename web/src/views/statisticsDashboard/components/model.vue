<template>
  <div id="statistics_client" class="statistics_common list-common">
    <div>
      <div style="padding: 5px 24px">
        <label>{{ $t('statisticsDashboard.modelSelect') }}:</label>
        <el-select
          v-model="modelParams.modelType"
          :placeholder="$t('modelAccess.table.modelType')"
          class="no-border-select"
          style="margin-left: 15px"
          clearable
          @change="fetchModels()"
        >
          <el-option
            v-for="item in modelTypeList"
            :key="item.key"
            :label="item.name"
            :value="item.key"
          />
        </el-select>
        <el-select
          v-model="modelParams.models"
          :placeholder="$t('statisticsDashboard.model')"
          class="no-border-select"
          style="margin-left: 15px; width: 600px"
          clearable
          multiple
          filterable
        >
          <el-option
            v-for="item in modelList"
            :key="item.modelId"
            :label="item.displayName || item.model"
            :value="item.modelId"
          />
        </el-select>
      </div>
      <div>
        <Search
          v-show="searchShow"
          ref="search"
          @handleSetTime="handleSetTime"
        ></Search>
      </div>
    </div>
    <div class="statistics_content_box">
      <div class="item_box">
        <div class="dataOverview">
          <span class="title">
            {{ $t('statistics.overview') }}
          </span>
          <div class="client_dataOverview_content" v-loading="loading">
            <div v-for="(item, index) in count" :key="index" class="card">
              <span>
                {{ item.name }}
                <i
                  :style="{
                    background: item.des_value < 0 ? '#1afa29' : '#d81e06',
                  }"
                  :class="{
                    defaultBg: item.des_value === 0 || item.des_value === -9999,
                  }"
                ></i>
              </span>
              <strong>{{ formatAmount(item.value) }}{{ item.unit }}</strong>
              <span>
                {{ item.des }}
                <label
                  :style="{
                    color: item.des_value < 0 ? '#1afa29' : '#d81e06',
                  }"
                  :class="{
                    defaultColor:
                      item.des_value === 0 || item.des_value === -9999,
                  }"
                >
                  {{ item.des_value === -9999 ? '-' : item.des_value + '%' }}
                </label>
                <img
                  v-if="item.des_value < 0 && item.des_value !== -9999"
                  src="@/assets/imgs/descend.png"
                  alt=""
                />
                <img
                  v-if="item.des_value > 0 && item.des_value !== -9999"
                  src="@/assets/imgs/rise.png"
                  alt=""
                />
              </span>
            </div>
          </div>
        </div>
        <div class="data_echart_box">
          <div class="data_echart">
            <UserEchart
              :content="
                echartContent.modelCalls ? echartContent.modelCalls.lines : []
              "
              :name="
                echartContent.modelCalls
                  ? echartContent.modelCalls.tableName
                  : ''
              "
              v-loading="loading"
            ></UserEchart>
          </div>
          <div class="data_echart">
            <UserEchart
              :content="
                echartContent.tokensUsage ? echartContent.tokensUsage.lines : []
              "
              :name="
                echartContent.tokensUsage
                  ? echartContent.tokensUsage.tableName
                  : ''
              "
              v-loading="loading"
            ></UserEchart>
          </div>
        </div>
        <div class="dataOverview">
          <span class="title">
            {{ $t('statisticsDashboard.modelList') }}
          </span>
          <div style="margin-top: -20px">
            <ModelList ref="modelList" />
          </div>
        </div>
      </div>
    </div>
    <el-backtop target=".statistics_content_box"></el-backtop>
  </div>
</template>
<script>
import Search from '@/components/searchDate.vue';
import UserEchart from '@/components/echart/userEchart.vue';
import ModelList from './modelList.vue';
import { formatAmount } from '@/utils/util.js';
import { getModelData } from '@/api/statisticsDashboard';
import { MODEL_TYPE } from '@/views/modelAccess/constants';
import { fetchModelList } from '@/api/modelAccess';

export default {
  components: {
    UserEchart,
    Search,
    ModelList,
  },
  data() {
    return {
      modelTypeList: MODEL_TYPE,
      modelList: [],
      loading: false,
      content: {}, // 存储返回的总揽数据
      echartContent: {}, // 存储返回的echart数据
      type: 'model',
      concurrentUser: {},
      count: [
        {
          name: this.$t('statisticsDashboard.tokenTotals'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'totalTokensTotal',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.quantity'),
        },
        {
          name: this.$t('statisticsDashboard.promptTokensTotals'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'promptTokensTotal',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.quantity'),
        },
        {
          name: this.$t('statisticsDashboard.completionTokensTotals'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'completionTokensTotal',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.quantity'),
        },
        {
          name: this.$t('statisticsDashboard.avgCosts'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'avgCosts',
          des_value: -9999,
          unit: 'ms',
        },
        {
          name: this.$t('statisticsDashboard.callCount'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'callCountTotal',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.callFailure'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'callFailureTotal',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.avgFirstTokenLatency'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'avgFirstTokenLatency',
          des_value: -9999,
          unit: 'ms',
        },
      ],
      searchShow: true,
      timeout: null, // 防抖定时
      searchTime: {
        time: [],
      },
      modelParams: {
        modelType: '',
        models: [],
      },
    };
  },
  computed: {
    params() {
      return {
        endDate: this.searchTime.time[1],
        startDate: this.searchTime.time[0],
      };
    },
  },
  methods: {
    formatAmount,
    formatParams(params) {
      return {
        ...params,
        models: params.models ? params.models.toString() : '',
      };
    },
    async fetchModels() {
      const res = await fetchModelList({
        filterScope: '',
        modelType: this.modelParams.modelType,
      });
      const modelList = res.data ? res.data.list || [] : [];
      this.modelList = modelList.filter(item => item.allowEdit);
    },
    handleSetTime(val) {
      this.loading = true;
      this.searchTime = val;

      const params = this.formatParams({
        startDate: val.time[0],
        endDate: val.time[1],
        ...this.modelParams,
      });
      getModelData(params)
        .then(res => {
          const { overview, trend } = res.data || {};
          this.content = overview || {};
          this.echartContent = trend || {};
          // 解构后台返回的数据，暂存和 count 数组中key对应的数据
          this.count.map(item => {
            item.value = overview[item.key] ? overview[item.key].value : 0;
            item.des_value = overview[item.key]
              ? overview[item.key].periodOverPeriod
              : -9999;
          });
        })
        .finally(() => {
          this.loading = false;
        });
      this.$refs.modelList.getTableData(params);
    },
  },
};
</script>
<style lang="scss" scoped>
.statistics_common {
  position: relative;
  height: 100%;
  padding: 0;
  /*background: #fff;*/
  overflow: hidden;
  z-index: 100;
  .statistics_content_box {
    position: relative;
    height: 100%;
    padding: 0 24px 0 24px;
    overflow-y: auto;
  }
  .el-radio-button__inner {
    cursor: pointer;
    color: $color;
  }
  .el-radio-button {
    &.is-active {
      span {
        color: #fff !important;
        background: $color !important;
      }
    }
    &.is-disabled {
      span {
        color: #999 !important;
        box-shadow: none;
      }
    }
  }
  .el-backtop {
    i {
      font-size: 20px;
      color: $color;
    }
  }
  .my-pagination {
    ::v-deep.el-pagination {
      text-align: right;
    }
    .el-pagination.is-background .el-pager li:not(.disabled).active {
      background-color: $color;
      color: #fff;
    }

    .el-pagination .el-select .el-input .el-input__inner {
      padding-right: 25px;
      width: 109px;
      border-color: #cccccc;
    }

    .el-pagination .el-select .el-input .el-input__inner {
      padding-right: 25px;
      width: 109px;
      border-color: #cccccc;
    }

    .el-pager li:hover {
      color: $color;
    }
    .el-pager li.active {
      color: $color;
      cursor: default;
      border: 1px solid $color;
    }

    .el-pagination__editor.el-input .el-input__inner {
      height: 28px;
      background: #ffffff;
      border: 1px solid #cccccc;
    }
    .el-pagination.is-background .el-pager li:not(.disabled):hover {
      color: $color;
    }
  }
  .el-empty {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    padding: 0;
  }
  .el-empty__image {
    width: 15%;
  }
}

#statistics_client {
  .item_box {
    .client_const {
      display: flex;
      padding: 20px 0;
      justify-content: space-between;
      background: #fff;
      margin-bottom: 20px;
      border-radius: 5px;

      span {
        display: flex;
        justify-content: center;
        align-items: center;
        width: calc(100% / 3);
        height: 100px;
        border-left: 1px solid #e8e9eb;

        &:first-child {
          border: 0;
        }

        img {
          height: 70px;
          margin-right: 20px;
        }

        div {
          display: flex;
          flex-direction: column;
          justify-content: space-between;
          height: 70px;
          font-size: 15px;

          strong {
            font-size: 20px;
          }
        }
      }
    }

    .defaultColor {
      color: #abb0b5 !important;
    }

    .defaultBg {
      background: #abb0b5 !important;
    }

    .data_echart_box {
      display: flex;
      justify-content: space-between;

      .data_echart {
        width: calc(50% - 10px);
      }
    }

    .data_echart {
      display: inline-block;
      width: 100%;
      position: relative;
      margin-bottom: 20px;
      background: #fff;
      border-radius: 5px;

      .title {
        display: block;
        margin-bottom: 20px;
      }

      .el-radio-group {
        margin-bottom: 20px;
      }
    }

    .dataOverview {
      padding: 20px;
      margin-bottom: 20px;
      margin-top: 15px;
      background: #fff;
      flex: 1;
      border-radius: 5px;
      box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.08);

      .client_dataOverview_content {
        display: flex;
        flex-wrap: wrap;
        justify-content: flex-start;
        margin-top: 20px;

        .card {
          position: relative;
          width: 24%;
          height: 120px;
          background: rgb(245, 246, 249);
          border-radius: 4px;
          padding: 15px 0;
          margin-left: 0.5%;
          margin-right: 0.5%;
          margin-bottom: 20px;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: space-between;

          &:nth-last-child(-n + 7) {
            margin-bottom: 0;
          }

          span {
            position: relative;

            &:last-child {
              display: flex;
              align-items: center;
              justify-content: space-between;
              font-size: 12px;
              color: rgb(171, 176, 181);
            }

            label {
              color: #303133;
              margin-left: 10px;
              font-weight: bold;
              font-size: 14px;
            }

            img {
              width: 13px;
              vertical-align: middle;
            }
          }

          strong {
            font-size: 15px;
          }

          i {
            position: absolute;
            width: 8px;
            height: 8px;
            border-radius: 50%;
            top: 5px;
            left: -13px;
            z-index: 1;
          }
        }
      }
    }

    .title {
      position: relative;
      font-size: 14px;
      font-weight: bold;
      padding-left: 10px;

      &::after {
        content: '';
        width: 3px;
        height: 15px;
        background: $color;
        position: absolute;
        left: 0;
        top: 50%;
        transform: translate(0, -50%);
      }

      label {
        font-size: 10px;
        color: rgb(171, 176, 181);
      }
    }
  }
}
</style>
