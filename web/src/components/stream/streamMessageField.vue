<!--问答消息框-->
<template>
  <div class="session rl">
    <div v-if="supportClear" class="session-setting">
      <el-link
        class="right-setting"
        @click="gropdownClick"
        type="primary"
        :underline="false"
        style="color: var(--color); top: 0"
      >
        <span class="el-icon-delete"></span>
        {{ $t('app.clearChat') }}
      </el-link>
    </div>
    <div
      class="history-box showScroll"
      :id="scrollContainerId"
      v-loading="loading"
      ref="timeScroll"
      @click="handleGlobalClick"
      :style="{ 'max-height': historyBoxHeight }"
    >
      <div v-for="(n, i) in session_data.history" :key="`${i}sdhs`">
        <!--问题-->
        <div v-if="n.query" class="session-question">
          <div :class="['session-item', 'rl']">
            <img class="logo" :src="userAvatarSrc" />
            <div class="answer-content">
              <div class="answer-content-query">
                <el-popover
                  placement="bottom-start"
                  trigger="hover"
                  :visible-arrow="false"
                  popper-class="query-copy-popover"
                  content=""
                >
                  <p
                    class="query-copy"
                    @click="queryCopy(n.query)"
                    style="cursor: pointer"
                  >
                    <i class="el-icon-s-order"></i>
                    &nbsp;
                    {{ $t('agent.copyToInput') }}
                  </p>
                  <span
                    slot="reference"
                    class="answer-text"
                    style="display: inline-block; margin-top: 5px"
                  >
                    {{ n.query }}
                  </span>
                </el-popover>
                <div class="echo-doc-box" v-if="hasFiles(n)">
                  <el-button
                    v-show="canScroll(i, n.showScrollBtn)"
                    icon="el-icon-arrow-left "
                    @click="prev($event, i)"
                    circle
                    class="scroll-btn left"
                    size="mini"
                    type="primary"
                  ></el-button>
                  <div class="imgList" :ref="`imgList-${i}`">
                    <div
                      v-for="(file, j) in n.fileList"
                      :key="`${j}sdsl`"
                      class="docInfo-img-container"
                    >
                      <img
                        v-if="hasImgs(n, file)"
                        :src="file.fileUrl"
                        class="docIcon imgIcon"
                      />
                      <div v-else class="docInfo-container">
                        <img
                          :src="require('@/assets/imgs/fileicon.png')"
                          class="docIcon"
                          style="width: 30px !important"
                        />
                        <div class="docInfo">
                          <p class="docInfo_name">
                            {{ $t('knowledgeManage.fileName') }}:{{ file.name }}
                          </p>
                          <p class="docInfo_size">
                            {{ $t('knowledgeManage.fileSize') }}:{{
                              getFileSizeDisplay(file.size)
                            }}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                  <el-button
                    v-show="canScroll(i, n.showScrollBtn)"
                    icon="el-icon-arrow-right"
                    @click="next($event, i)"
                    circle
                    class="scroll-btn right"
                    size="mini"
                    type="primary"
                  ></el-button>
                </div>
              </div>
            </div>
          </div>
        </div>
        <!--loading-->
        <div v-if="n.responseLoading" class="session-answer">
          <div class="session-answer-wrapper">
            <img class="logo" :src="modelIconUrl || avatarSrc(defaultUrl)" />
            <div class="answer-content"><i class="el-icon-loading"></i></div>
          </div>
        </div>
        <!--pending-->
        <div v-if="n.pendingResponse" class="session-answer">
          <div class="session-answer-wrapper">
            <img class="logo" :src="modelIconUrl || avatarSrc(defaultUrl)" />
            <div class="answer-content" style="padding: 10px; color: #e6a23c">
              {{ n.pendingResponse }}
            </div>
          </div>
        </div>

        <!-- 回答故障  code:7-->
        <div class="session-error" v-if="n.error">
          <i class="el-icon-warning"></i>
          &nbsp;{{ n.response }}
        </div>

        <!--回答 文字+图片-->
        <div
          v-if="
            !n.error &&
            (n.response ||
              n.msg_type ||
              (n.subConversions && n.subConversions.length))
          "
          class="session-answer"
          :id="'message-container' + i"
        >
          <!-- v-if="[0].includes(n.qa_type)" -->
          <div class="session-answer-wrapper">
            <img class="logo" :src="modelIconUrl || avatarSrc(defaultUrl)" />
            <div class="session-wrap" style="width: calc(100% - 30px)">
              <!-- 思考块显示 (msg_type 逻辑) -->
              <div
                class="deepseek"
                v-if="
                  n.msg_type &&
                  ['qa_start', 'qa_finish', 'knowledge_start'].includes(
                    n.msg_type,
                  )
                "
              >
                <img
                  :src="require('@/assets/imgs/think-icon.png')"
                  class="think_icon"
                />
                {{ getTitle(n.msg_type) }}
              </div>
              <div v-else-if="chatType === 'rag'">
                <img
                  :src="require('@/assets/imgs/think-icon.png')"
                  class="think_icon"
                />
                <div
                  v-if="showDSBtn(n.response)"
                  class="deepseek"
                  @click="toggle($event, i)"
                >
                  {{ n.thinkText }}
                  <i
                    v-bind:class="{
                      'el-icon-arrow-down': !n.isOpen,
                      'el-icon-arrow-up': n.isOpen,
                    }"
                  ></i>
                </div>
                <span v-else class="deepseek">
                  {{ $t('menu.knowledge') }}
                </span>
              </div>
              <div
                v-else-if="
                  !(n.messageSequence && n.messageSequence.length) &&
                  showDSBtn(n.response)
                "
              >
                <div class="deepseek" @click="toggle($event, i)">
                  <img
                    :src="require('@/assets/imgs/think-icon.png')"
                    class="think_icon"
                  />
                  {{ n.thinkText }}
                  <i
                    v-bind:class="{
                      'el-icon-arrow-down': !n.isOpen,
                      'el-icon-arrow-up': n.isOpen,
                    }"
                  ></i>
                </div>
              </div>

              <!-- 消息序列渲染 -->
              <div
                v-if="n.messageSequence && n.messageSequence.length"
                class="message-sequence-wrapper"
              >
                <template v-for="(item, idx) in n.messageSequence">
                  <!-- 子会话渲染块 -->
                  <div
                    v-if="item.type === 'sub'"
                    :key="'sub-' + item.id + idx"
                    class="sub-conversion-box order-sub"
                  >
                    <sub-conversion
                      :conversion="findSubData(n, item.id)"
                      :parents-index="i"
                      @toggle-conversion="toggleSubConversion"
                      @collapse-click="collapseClick"
                    ></sub-conversion>
                  </div>

                  <!-- 主会话渲染块 -->
                  <div
                    v-else-if="item.type === 'main'"
                    :key="'main-' + idx"
                    class="order-main-chunk"
                  >
                    <!-- 片段内的思考按钮 -->
                    <div
                      v-if="
                        showDSBtn(
                          item.renderedContent || item.activeResponse || '',
                        )
                      "
                    >
                      <div class="deepseek" @click="toggle($event, i)">
                        <img
                          :src="require('@/assets/imgs/think-icon.png')"
                          class="think_icon"
                        />
                        {{ n.thinkText }}
                        <i
                          v-bind:class="{
                            'el-icon-arrow-down': !n.isOpen,
                            'el-icon-arrow-up': n.isOpen,
                          }"
                        ></i>
                      </div>
                    </div>
                    <template
                      v-if="
                        (item.stableChunks && item.stableChunks.length) ||
                        item.activeResponse
                      "
                    >
                      <div class="answer-content">
                        <div
                          v-for="(chunk, cIdx) in item.stableChunks"
                          :key="'stable-' + cIdx"
                          class="chunk_stable"
                          v-bind:class="{ 'ds-res': showDSBtn(chunk) }"
                          v-html="
                            showDSBtn(chunk) ? replaceHTML(chunk, n) : chunk
                          "
                        ></div>
                        <div
                          v-if="item.activeResponse"
                          class="chunk_active"
                          v-bind:class="{
                            'ds-res': showDSBtn(item.activeResponse),
                          }"
                          v-html="
                            showDSBtn(item.activeResponse)
                              ? replaceHTML(item.activeResponse, n)
                              : item.activeResponse
                          "
                        ></div>
                      </div>
                    </template>
                    <!-- 历史内容 -->
                    <div
                      v-else-if="item.renderedContent"
                      class="answer-content order-main-renderedContent"
                      v-bind:class="{
                        'ds-res': showDSBtn(item.renderedContent),
                        hideDs: !n.isOpen,
                      }"
                      v-html="
                        showDSBtn(item.renderedContent)
                          ? replaceHTML(item.renderedContent, n)
                          : item.renderedContent
                      "
                    ></div>
                  </div>
                </template>
              </div>

              <!-- 如果没有 messageSequence(rag) -->
              <template v-else>
                <!-- 子会话渲染区域 -->
                <div
                  v-if="n.subConversions && n.subConversions.length"
                  class="sub-conversion-box"
                >
                  <sub-conversion
                    v-for="(conversion, idx) in n.subConversions"
                    :key="idx"
                    :conversion="conversion"
                    :parents-index="i"
                    @toggle-conversion="toggleSubConversion"
                    @collapse-click="collapseClick"
                  ></sub-conversion>
                </div>

                <!-- 主会话-->
                <template
                  v-if="
                    (n.stableChunks && n.stableChunks.length) ||
                    n.activeResponse
                  "
                >
                  <div class="answer-content no-order-chunk-answer">
                    <div
                      v-for="(chunk, idx) in n.stableChunks"
                      :key="idx"
                      class="chunk_stable"
                      v-bind:class="{ 'ds-res': showDSBtn(chunk) }"
                      v-html="showDSBtn(chunk) ? replaceHTML(chunk, n) : chunk"
                    ></div>
                    <div
                      v-if="n.activeResponse"
                      class="chunk_active"
                      v-bind:class="{ 'ds-res': showDSBtn(n.activeResponse) }"
                      v-html="
                        showDSBtn(n.activeResponse)
                          ? replaceHTML(n.activeResponse, n)
                          : n.activeResponse
                      "
                    ></div>
                  </div>
                </template>
                <div
                  v-else-if="n.response"
                  class="answer-content history-answer"
                  v-bind:class="{
                    'ds-res': showDSBtn(n.response),
                    hideDs: !n.isOpen,
                  }"
                  v-html="
                    showDSBtn(n.response)
                      ? replaceHTML(n.response, n)
                      : n.response
                  "
                ></div>
              </template>
            </div>
          </div>
          <!-- <div v-else class="session-answer-wrapper">
            <img class="logo" :src="avatarSrc(defaultUrl)" />
            <div v-if="n.code === 7" class="answer-content session-error">
              <i class="el-icon-warning"></i>
              &nbsp;{{ n.response }}
            </div>
            <div v-else class="answer-content" v-html="n.response"></div>
          </div> -->
          <!--文件-->
          <div
            v-if="n.gen_file_url_list && n.gen_file_url_list.length"
            class="file-path response-file"
          >
            <el-image
              v-for="(g, k) in n.gen_file_url_list"
              :key="k"
              :src="g"
              :preview-src-list="[g]"
            ></el-image>
          </div>
          <!--出处-->
          <div
            v-if="n.searchList && n.searchList.length && n.finish === 1"
            class="search-list"
          >
            <h2
              class="recommended-question-title"
              v-if="n.msg_type && ['qa_finish'].includes(n.msg_type)"
            >
              {{ $t('app.recommendedQuestion') }}
            </h2>
            <div
              v-for="(m, j) in n.searchList"
              :key="`${j}sdsl`"
              class="search-list-item"
            >
              <div
                v-if="m.content_type && m.content_type === 'qa'"
                class="qa_content"
                @click="handleRecommendedQuestion(m)"
              >
                <span>{{ j + 1 }}. {{ m.question }}</span>
              </div>
              <template v-else>
                <div
                  class="serach-list-item"
                  v-if="showSearchList(j, n.citations)"
                >
                  <span @click="collapseClick(n, m, j)">
                    <i
                      :class="[
                        '',
                        m.collapse
                          ? 'el-icon-caret-bottom'
                          : 'el-icon-caret-right',
                      ]"
                    ></i>
                    {{ $t('agent.source') }}：
                  </span>
                  <a
                    v-if="m.link"
                    :href="m.link"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="link"
                  >
                    {{ m.link }}
                  </a>
                  <span v-if="m.title">
                    <sub
                      class="subTag"
                      :data-parents-index="i"
                      :data-collapse="m.collapse ? 'true' : 'false'"
                    >
                      {{ j + 1 }}
                    </sub>
                    {{ m.title }}
                  </span>
                  <!-- <span @click="goPreview($event,m)" class="search-doc">查看全文</span> -->
                </div>
                <el-collapse-transition>
                  <div v-show="m.collapse ? true : false" class="snippet">
                    <p v-html="m.snippet"></p>
                  </div>
                </el-collapse-transition>
              </template>
            </div>
          </div>
          <!--loading-->
          <div
            v-if="
              n.finish === 0 &&
              sessionStatus == 0 &&
              i === session_data.history.length - 1
            "
            class="text-loading"
          >
            <div></div>
            <div></div>
            <div></div>
          </div>
          <!--停止生成 重新生成 点赞   session code 是0时不可操作-->
          <div class="answer-operation">
            <div class="opera-left">
              <span
                v-if="
                  i === session_data.history.length - 1 && sessionStatus !== 0
                "
                class="restart"
                @click="refresh"
              >
                <img :src="require('@/assets/imgs/refresh-icon.png')" />
              </span>
              <span
                class="preStop"
                @click="preStop"
                v-if="
                  supportStop &&
                  i === session_data.history.length - 1 &&
                  sessionStatus === 0
                "
              >
                <img :src="require('@/assets/imgs/stop-icon.png')" />
              </span>
            </div>
            <div
              class="opera-right"
              style="flex: 0"
              @click="
                () => {
                  copy(n.oriResponse) && copycb();
                }
              "
            >
              <img :src="require('@/assets/imgs/copy-icon.png')" />
            </div>
            <!--提示话术-->
            <div class="answer-operation-tip">
              {{ $t('agent.answerOperationTip') }}
            </div>
          </div>
          <!-- 推荐问题 -仅最后一条回答显示 -->
          <div
            v-if="
              sessionStatus === -1 &&
              ((recommendConfig.list && recommendConfig.list.length) ||
                recommendConfig.loading) &&
              i === session_data.history.length - 1
            "
            class="session-section-wrapper recommend-question"
          >
            <div
              v-for="(item, index) in recommendConfig.list"
              :key="index"
              :class="[
                'recommend-question-item',
                { 'is-tips': item.type === 'tips' },
              ]"
              @click="
                item.type !== 'tips' &&
                $emit('handleRecommendClick', item.content)
              "
            >
              {{ item.content }}
            </div>
            <div
              v-if="recommendConfig.loading"
              class="text-loading recommend-question-loading"
            >
              <div></div>
              <div></div>
              <div></div>
            </div>
          </div>
        </div>

        <!-- 回答 仅图片-->
        <div
          v-if="
            !n.response && n.gen_file_url_list && n.gen_file_url_list.length
          "
          class="session-answer"
        >
          <div class="session-answer-wrapper">
            <img class="logo" :src="modelIconUrl || avatarSrc(defaultUrl)" />
            <div class="answer-content">
              <div
                v-if="n.gen_file_url_list && n.gen_file_url_list.length"
                class="file-path response-file no-response"
              >
                <el-image
                  v-for="(g, k) in n.gen_file_url_list"
                  :key="k"
                  :src="g"
                  :preview-src-list="[g]"
                ></el-image>
              </div>
            </div>
          </div>
          <!--仅图片时只有 重新生成-->
          <div class="answer-operation">
            <div class="opera-left">
              <span
                v-if="i === session_data.history.length - 1"
                class="restart"
              >
                <i class="el-icon-refresh" @click="refresh">
                  &nbsp;
                  {{ $t('agent.refresh') }}
                </i>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import smoothscroll from 'smoothscroll-polyfill';
import { md } from '@/mixins/markdown-it';
import { marked } from 'marked';
var highlight = require('highlight.js');
import 'highlight.js/styles/atom-one-dark.css';
import commonMixin from '@/mixins/common';
import { mapGetters, mapState } from 'vuex';
import { avatarSrc } from '@/utils/util';
import SubConversion from './subConversion.vue';

marked.setOptions({
  renderer: new marked.Renderer(),
  gfm: true,
  tables: true,
  breaks: false,
  pedantic: false,
  sanitize: false,
  smartLists: true,
  smartypants: false,
  highlight: function (code) {
    return highlight.highlightAuto(code).value;
  },
});

export default {
  mixins: [commonMixin],
  props: {
    defaultUrl: {
      type: String,
      default: '',
    },
    chatType: {
      type: String,
      default: '',
    },
    recommendConfig: {
      type: Object,
      default: () => ({
        reqController: null,
        list: [],
        loading: false,
      }),
    },
    modelIconUrl: {},
    supportStop: {},
    modelSessionStatus: {},
    supportClear: {
      type: Boolean,
      default: true,
    },
  },
  components: {
    SubConversion,
  },
  data() {
    return {
      md: md,
      autoScroll: true,
      scrollTimeout: null,
      loading: false,
      marked: marked,
      session_data: {
        tool: '',
        searchList: [],
        history: [],
        response: '',
      },
      c: null,
      ctx: null,
      canvasShow: false,
      cv: null,
      currImg: {
        url: '',
        width: 0, // 原始宽高
        height: 0,
        w: 0, // 压缩后的宽高
        h: 358,
        roteX: 0, // 压缩后的比例
        roteY: 0,
      },
      imgConfig: ['jpeg', 'PNG', 'png', 'JPG', 'jpg', 'bmp', 'webp'],
      audioConfig: ['mp3', 'wav'],
      fileScrollStateMap: {},
      resizeTimer: null,
      scrollContainerId: `timeScroll-${this._uid}`,
      // 复制提示计时器map
      copyTimerMap: new Map(),
      historyBoxHeight: '', // 动态历史会话容器高度
    };
  },
  computed: {
    ...mapGetters('user', ['userAvatar']),
    // ...mapState('app', ['sessionStatus']),
    sessionStatus() {
      return ['number', 'string'].includes(typeof this.modelSessionStatus)
        ? this.modelSessionStatus
        : this.$store.state.app.sessionStatus;
    },
    userAvatarSrc() {
      return this.userAvatar
        ? avatarSrc(this.userAvatar)
        : require('@/assets/imgs/robot-icon.png');
    },
    isStreaming() {
      const history = this.session_data.history;
      if (history.length === 0) return false;
      const lastItem = history[history.length - 1];
      return lastItem.finish === 0 && this.sessionStatus === 0;
    },
  },
  watch: {
    'session_data.history': {
      handler() {
        this.$nextTick(() => {
          this.updateAllFileScrollStates();
        });
      },
      deep: true,
    },
  },
  mounted() {
    this.setupScrollListener();
    smoothscroll.polyfill();
    document.addEventListener('click', this.handleCitationClick);
    document.addEventListener('click', this.handleCitationBtnClick);
    window.addEventListener('resize', this.handleWindowResize);
    this.updateAllFileScrollStates();
  },
  beforeDestroy() {
    if (this.handleCitationClick) {
      document.removeEventListener('click', this.handleCitationClick);
    }
    if (this.handleCitationBtnClick) {
      document.removeEventListener('click', this.handleCitationBtnClick);
    }
    const container = document.getElementById(this.scrollContainerId);
    if (container) {
      container.removeEventListener('scroll', this.handleScroll);
    }
    clearTimeout(this.scrollTimeout);

    window.removeEventListener('resize', this.handleWindowResize);
    if (this.resizeTimer) {
      clearTimeout(this.resizeTimer);
    }
    // 移除图片错误事件监听器
    if (this.imageErrorHandler) {
      document.body.removeEventListener('error', this.imageErrorHandler, true);
    }
    // 清除复制提示计时器
    this.copyTimerMap.forEach(timerId => {
      clearTimeout(timerId);
    });
    this.copyTimerMap.clear();
  },
  methods: {
    avatarSrc,
    getTitle(type) {
      if (type === 'qa_start') {
        return this.$t('app.qaSearching');
      } else if (type === 'knowledge_start') {
        return this.$t('app.knowledgeSearch');
      } else if (type === 'qa_finish') {
        return this.$t('knowledgeManage.qaDatabase.title');
      } else {
        return this.$t('menu.knowledge');
      }
    },
    handleRecommendedQuestion(m) {
      this.$emit('handleRecommendedQuestion', m.question);
    },
    handleCitationBtnClick(e) {
      const target = e.target;
      if (target.classList.contains('citation-tips-content-icon')) {
        const index = target.dataset.index;
        const citation = Number(target.dataset.citation);
        const historyItem = this.session_data.history[index];
        if (historyItem && historyItem.searchList) {
          const searchItem = historyItem.searchList[citation - 1];
          if (searchItem) {
            const j = historyItem.searchList.indexOf(searchItem);
            this.collapseClick(historyItem, searchItem, j);
          }
        }
      }
    },
    updateAllFileScrollStates() {
      this.session_data.history.forEach((item, index) => {
        if (item.fileList && item.fileList.length > 0) {
          this.$nextTick(() => {
            this.checkFileScrollState(index);
          });
        }
      });
    },
    checkFileScrollState(index) {
      const refKey = `imgList-${index}`;
      const containerArray = this.$refs[refKey];
      if (containerArray && containerArray.length > 0) {
        const container = containerArray[0];
        const canScroll = container.scrollWidth > container.clientWidth;
        if (this.session_data.history[index]) {
          this.$set(
            this.session_data.history[index],
            'showScrollBtn',
            canScroll,
          );
        }
        this.$set(this.fileScrollStateMap, index, canScroll);
      }
    },
    handleWindowResize() {
      if (this.resizeTimer) {
        clearTimeout(this.resizeTimer);
      }
      this.resizeTimer = setTimeout(() => {
        this.updateAllFileScrollStates();
      }, 200);
    },
    canScroll(i, showScrollBtn) {
      if (showScrollBtn !== null && showScrollBtn !== undefined) {
        return showScrollBtn;
      }
      return this.fileScrollStateMap[i] || false;
    },
    prev(e, i) {
      e.stopPropagation();
      const refKey = `imgList-${i}`;
      const containerArray = this.$refs[refKey];
      if (containerArray && containerArray.length > 0) {
        const container = containerArray[0];
        container.scrollBy({
          left: -200,
          behavior: 'smooth',
        });
      }
    },
    next(e, i) {
      e.stopPropagation();
      const refKey = `imgList-${i}`;
      const containerArray = this.$refs[refKey];
      if (containerArray && containerArray.length > 0) {
        const container = containerArray[0];
        container.scrollBy({
          left: 200,
          behavior: 'smooth',
        });
      }
    },
    hasFiles(n) {
      return n.fileList && n.fileList.length > 0;
    },
    hasImgs(n, file) {
      if (!n.fileList || n.fileList.length === 0 || !file || !file.name) {
        return false;
      }
      let type = file.name.split('.').pop().toLowerCase();
      return this.imgConfig.map(t => t.toLowerCase()).includes(type);
    },
    handleCitationClick(e) {
      const target = e.target.closest('.citation');
      const subConversionItem = target
        ? target.closest('.sub-conversion-item')
        : null;

      if (subConversionItem && target.dataset.pid) {
        // 子会话引用点击处理
        const pid = target.dataset.pid;
        // 父会话在 history 中的索引
        const parentsIndex = Number(target.dataset.parentsIndex);
        // 引用tag
        const citationIndex = Number(target.textContent);
        const historyItem = this.session_data.history[parentsIndex];

        if (historyItem && historyItem.subConversions) {
          // 在对应的 history 项中找到该子会话实例
          const subConversion = historyItem.subConversions.find(
            a => a.id === pid,
          );

          if (
            subConversion &&
            subConversion.searchList &&
            subConversion.searchList[citationIndex - 1]
          ) {
            const searchItem = subConversion.searchList[citationIndex - 1];
            this.collapseClick(subConversion, searchItem, citationIndex - 1);

            // 视图滚动：在当前子会话容器范围内寻找对应的引用详情项
            this.$nextTick(() => {
              const targetSearchItem = subConversionItem.querySelector(
                `.search-list-item[data-citation-index="${citationIndex}"]`,
              );
              if (targetSearchItem) {
                targetSearchItem.scrollIntoView({
                  behavior: 'smooth',
                  block: 'center',
                });
              }
            });

            e.stopPropagation();
            return;
          }
        }
      }

      this.$handleCitationClick(e, {
        sessionStatus: this.sessionStatus,
        sessionData: this.session_data,
        citationSelector: '.citation',
        scrollElementId: this.scrollContainerId,
        onToggleCollapse: (item, collapse) => {
          this.$set(item, 'collapse', collapse);
        },
      });
    },
    // 子会话展开收起
    toggleSubConversion(conversion) {
      this.$set(conversion, 'isOpen', !conversion.isOpen);
    },
    showSearchList(j, citations) {
      return (citations || []).includes(j + 1);
    },
    setCitations(index) {
      let citation = `#message-container${index} .citation`;
      const allCitations = document.querySelectorAll(citation);
      const citationsSet = new Set();

      allCitations.forEach(element => {
        const text = element.textContent.trim();
        if (text) {
          citationsSet.add(Number(text));
        }
      });

      return Array.from(citationsSet);
    },
    goPreview(event, item) {
      event.stopPropagation();
      let { meta_data } = item;
      let { file_name, download_link, page_num, row_num, sheet_name } =
        meta_data;
      var index = file_name.lastIndexOf('.');
      var ext = file_name.substr(index + 1);
      let openUrl = '';
      let fileUrl = encodeURIComponent(download_link);
      const fileType = ['docx', 'doc', 'txt', 'pdf', 'xlsx'];
      if (fileType.includes(ext)) {
        switch (ext) {
          case 'docx' || 'doc':
            openUrl = `${window.location.origin}/doc?fileUrl=` + fileUrl;
            break;
          case 'txt':
            openUrl = `${window.location.origin}/txtView?fileUrl=` + fileUrl;
            break;
          case 'pdf':
            if (page_num.length > 0) {
              openUrl =
                `${window.location.origin}/pdfView?fileUrl=` +
                fileUrl +
                '&page=' +
                page_num[0];
            }
            break;
          case 'xlsx':
            openUrl =
              `${window.location.origin}/jsExcel?url=` +
              fileUrl +
              '&rownum=' +
              row_num +
              '&sheetName=' +
              sheet_name;
            break;
          default:
            this.$message.warning('暂不支持此格式查看');
        }
      }
      if (openUrl !== '') {
        window.open(openUrl, '_blank', 'noopener,noreferrer');
      } else {
        this.$message.warning('暂不支持此格式查看');
      }
    },
    setupScrollListener() {
      const container = document.getElementById(this.scrollContainerId);
      container.addEventListener('scroll', this.handleScroll);
    },
    handleScroll(e) {
      const container = document.getElementById(this.scrollContainerId);
      const { scrollTop, clientHeight, scrollHeight } = container;
      const nearBottom = scrollHeight - (scrollTop + clientHeight) < 5;
      if (!nearBottom) {
        this.autoScroll = false;
      }
      clearTimeout(this.scrollTimeout);
      this.scrollTimeout = setTimeout(() => {
        if (nearBottom) {
          this.autoScroll = true;
          this.scrollBottom();
        }
      }, 500);
    },
    replaceHTML(data, n) {
      const thinkStart = /<think>/i;
      const thinkEnd = /<\/think>/i;
      const toolStart = /<tool>/i;
      const toolEnd = /<\/tool>/i;

      // 处理 think 标签
      if (thinkEnd.test(data)) {
        // n.thinkText = '已深度思考';
        n.thinkText = this.$t('agent.thinked');
        if (!thinkStart.test(data)) {
          data = '<think>\n' + data;
        }
      }

      // 新增处理 tool 标签
      if (toolEnd.test(data)) {
        // n.toolText = '已使用工具';
        n.thinkText = this.$t('agent.thinked');
        if (!toolStart.test(data)) {
          data = '<tool>\n' + data;
        }
      }

      // 统一替换为 section 标签
      return data
        .replace(/think>/gi, 'section>')
        .replace(/tool>/gi, 'section>');
    },
    showDSBtn(data) {
      const pattern = /<(think|tool)(\s[^>]*)?>|<\/(think|tool)>/;
      const matches = data.match(pattern);
      if (!matches) {
        return false;
      }
      return true;
    },
    toggle(event, index) {
      const name = event.target.className;
      if (
        name === 'deepseek' ||
        name === 'el-icon-arrow-up' ||
        name === 'el-icon-arrow-down'
      ) {
        this.session_data.history[index].isOpen =
          !this.session_data.history[index].isOpen;
        this.$set(
          this.session_data.history,
          index,
          this.session_data.history[index],
        );
      }
    },
    queryCopy(text) {
      this.$emit('queryCopy', text);
    },
    getSessionData() {
      return this.session_data;
    },
    copy(text) {
      text = text.replaceAll('<br/>', '\n');
      var textareaEl = document.createElement('textarea');
      textareaEl.setAttribute('readonly', 'readonly');
      textareaEl.value = text;
      document.body.appendChild(textareaEl);
      textareaEl.select();
      var res = document.execCommand('copy');
      document.body.removeChild(textareaEl);
      return res;
    },
    copycb() {
      this.$message.success(this.$t('agent.copyTips'));
    },
    /**
     * 处理出处列表的展开/折叠点击事件
     * @param {Object} sourceContainer - 包含出处列表的容器对象（如主历史项或子会话对象）
     * @param {Object} searchItem - 当前点击的出处条目对象
     * @param {number} index - 当前条目在 searchList 中的索引
     */
    collapseClick(sourceContainer, searchItem, index) {
      this.$set(sourceContainer.searchList, index, {
        ...searchItem,
        collapse: !searchItem.collapse,
      });
    },
    doLoading() {
      this.loading = true;
    },
    scrollBottom() {
      this.loading = false;
      if (!this.autoScroll) return;
      this.$nextTick(() => {
        document.getElementById(this.scrollContainerId).scrollTop =
          document.getElementById(this.scrollContainerId).scrollHeight;
      });
    },

    codeScrollBottom() {
      this.$nextTick(() => {
        this.loading = false;
        document.getElementsByTagName('code').scrollTop =
          document.getElementsByTagName('code').scrollHeight;
      });
    },
    pushHistory(data) {
      this.session_data.history.push(data);
      this.scrollBottom();
    },
    replaceLastData(index, data) {
      if (!data.response && data.finish !== 0) {
        data.response = this.$t('app.noResponse');
      }
      this.$set(this.session_data.history, index, data);
      this.scrollBottom();
      this.codeScrollBottom();
      if (data.finish === 1) {
        this.$nextTick(() => {
          const setCitations = this.setCitations(index);
          this.$set(
            this.session_data.history[index],
            'citations',
            setCitations,
          );
        });
      }
    },
    getFileSizeDisplay(fileSize) {
      if (!fileSize || typeof fileSize !== 'number' || isNaN(fileSize)) {
        return '...';
      }
      return fileSize > 1024
        ? `${(fileSize / (1024 * 1024)).toFixed(2)} MB`
        : `${fileSize} bytes`;
    },
    replaceData(data) {
      this.session_data = data;
      this.scrollBottom();
    },
    replaceHistory(data) {
      this.session_data.history = data;
      this.$nextTick(() => {
        this.session_data.history.forEach((n, index) => {
          const setCitations = this.setCitations(index);
          this.$set(
            this.session_data.history[index],
            'citations',
            setCitations,
          );
        });
        this.scrollBottom();
      });
    },
    removeLastHistory() {
      this.session_data.history.pop();
    },
    replaceHistoryWithImg(data) {
      this.session_data.history = data;
      this.$nextTick(() => {
        this.preTagging(data[0].annotation);
      });
    },
    clearData() {
      this.session_data = {
        tool: '',
        searchList: [],
        history: [],
        response: '',
      };
    },
    loadAllImg() {
      this.session_data.history.forEach((n, i) => {
        n.gen_file_url_list.forEach((m, j) => {
          setTimeout(() => {
            this.$set(this.session_data.history[i].gen_file_url_list, j, {
              ...m,
              loadedUrl: m.url,
              loading: false,
            });
          }, 2000);
        });
      });
    },
    gropdownClick() {
      this.$emit('clearHistory');
    },
    getList() {
      return JSON.parse(
        JSON.stringify(
          this.session_data.history.filter(item => {
            delete item.operation;
            return item;
          }),
        ),
      );
    },
    getAllList() {
      return JSON.parse(JSON.stringify(this.session_data.history));
    },
    stopLoading() {
      this.session_data.history = this.session_data.history.filter(item => {
        return !item.pending;
      });
    },
    stopPending() {
      this.session_data.history = this.session_data.history.filter(item => {
        if (item.pending) {
          item.responseLoading = false;
          item.pendingResponse = this.$t('app.stopStream');
        }
        return item;
      });
    },
    refresh() {
      if (this.sessionStatus === 0) {
        return;
      }
      this.$emit('refresh');
    },
    preStop() {
      if (this.sessionStatus === 0) {
        this.$emit('preStop');
      }
    },
    preZan(index, item) {
      if (this.sessionStatus === 0) {
        return;
      }
      this.$set(this.session_data.history, index, { ...item, evaluate: 1 });
    },
    preCai(index, item) {
      if (this.sessionStatus === 0) {
        return;
      }
      this.$set(this.session_data.history, index, { ...item, evaluate: 2 });
    },
    initCanvasUtil() {
      this.canvasShow = true;
      this.$nextTick(() => {
        this.cv &&
          this.cv.destroy() &&
          this.cv.clearPre() &&
          this.cv.clearLabels() &&
          (this.cv = null);
        this.cv = new CanvasUtil(this);
      });
    },
    preTagging(response) {
      this.currImg = {
        url: '',
        width: 0,
        height: 0,
        w: 0,
        h: 358,
        roteX: 0,
        roteY: 0,
        dx: 0,
        dy: 0,
      };
      var image = new Image();
      image.src = response.annotationImg;
      image.onload = () => {
        this.currImg.width = image.width;
        this.currImg.height = image.height;
        this.c = document.getElementById('mycanvas');
        this.ctx = this.c.getContext('2d');
        this.resizeCanvas();
        this.initCanvasUtil();

        this.$nextTick(() => {
          this.echoLabels(response);
        });
      };
    },
    echoLabels(response) {
      this.cv.echoLabels(response);
    },
    resizeCanvas() {
      this.currImg.w = 0;
      this.currImg.h = 358;
      this.currImg.dx = 0;
      this.currImg.dy = 0;
      this.currImg.roteX = 0;
      this.currImg.roteY = 0;

      let currImg = this.currImg;
      let contain = document.getElementById('mycantain');
      if (currImg.width > contain.offsetWidth) {
        this.currImg.roteX = currImg.width / contain.offsetWidth;
        currImg.w = contain.offsetWidth;
        currImg.h = (currImg.height * contain.offsetWidth) / currImg.width;
        if (currImg.h > contain.offsetHeight) {
          currImg.h = contain.offsetHeight;
          currImg.w = (currImg.width * currImg.h) / currImg.height;
          currImg.roteX = currImg.width / currImg.w;
          currImg.dx = (contain.offsetWidth - currImg.w) / 2;
        } else {
          currImg.roteY = currImg.height / currImg.h;
          currImg.dy = (contain.offsetHeight - currImg.h) / 2;
        }
      } else {
        currImg.roteY = currImg.height / currImg.h;
        currImg.w = (currImg.width * currImg.h) / currImg.height;
        currImg.roteX = currImg.width / currImg.w;
        currImg.dx = (contain.offsetWidth - currImg.w) / 2;
      }

      this.canvasShow = true;
      this.c.width = currImg.w;
      this.c.height = currImg.h;
      this.$nextTick(() => {
        this.cv && this.cv.resizeCurrImg(currImg);
      });
    },
    // 初始化history列表
    initHistoryList(list) {
      this.$set(this.session_data, 'history', list);
      this.$nextTick(() => {
        this.updateAllFileScrollStates();
      });
    },
    handleGlobalClick(e) {
      // 复制
      if (e.target.classList.contains('copy-btn')) {
        const btn = e.target;
        if (this.copyTimerMap.has(btn)) {
          clearTimeout(this.copyTimerMap.get(btn));
        }
        let innerText = btn.parentNode.nextElementSibling.innerText;
        this.copy(innerText);
        this.$message.success(this.$t('agent.copyTips'));
        btn.innerText = this.$t('agent.copySuccess');
        const timerId = setTimeout(() => {
          btn.innerText = this.$t('agent.copy');
          this.copyTimerMap.delete(btn);
        }, 1500);
        this.copyTimerMap.set(btn, timerId);
      }
    },
    // 获取子会话数据
    findSubData(n, id) {
      if (!n.subConversions) return null;
      return n.subConversions.find(sub => sub.id === id);
    },
    // 动态设置滚动容器高度
    setHistoryBoxHeight(inputHeight) {
      if (inputHeight) {
        const baseInputHeight = 56;
        const offset = Math.max(0, inputHeight - baseInputHeight);
        this.historyBoxHeight = `calc(100% - ${46 + offset}px)`;
      } else {
        this.historyBoxHeight = '';
      }
      this.scrollBottom();
    },
  },
};
</script>

<style scoped lang="scss">
.serach-list-item {
  .link:hover {
    color: $color !important;
  }
  .search-doc {
    margin-left: 10px;
    cursor: pointer;
    color: $color !important;
  }
  .subTag {
    display: inline-flex;
    color: $color;
    border-radius: 50%;
    width: 18px;
    height: 18px;
    border: 1px solid $color;
    line-height: 18px;
    vertical-align: middle;
    margin-left: 2px;
    justify-content: center;
    align-items: center;
    font-size: 14px;
    overflow: hidden;
    white-space: nowrap;
    margin-bottom: 2px;
    transform: scale(0.8);
    font-style: normal;
  }
}

::v-deep {
  pre {
    white-space: pre-wrap !important;
    min-height: 50px;
    word-wrap: break-word;
    resize: vertical;
    .hljs {
      max-height: 300px !important;
      white-space: pre-wrap !important;
      min-height: 50px;
      word-wrap: break-word;
      resize: vertical;
    }
    code {
      display: block;
      white-space: pre-wrap;
      word-break: break-all;
      scroll-behavior: smooth;
    }
  }
  .el-loading-mask {
    background: none !important;
  }
  .answer-content {
    width: 100%;
    img {
      width: 80% !important;
    }
    section li,
    li {
      list-style-position: inside !important; /* 将标记符号放在内容框内 */
    }

    .citation {
      display: inline-flex;
      color: $color;
      border-radius: 50%;
      width: 18px;
      height: 18px;
      border: 1px solid $color;
      cursor: pointer;
      line-height: 18px;
      vertical-align: middle;
      margin-left: 5px;
      justify-content: center;
      align-items: center;
      font-size: 14px;
      overflow: hidden;
      white-space: nowrap;
      margin-bottom: 2px;
      transform: scale(0.8);
    }
  }
  .search-list {
    img {
      width: 80% !important;
    }
  }
}
.more {
  color: $color;
}
.session {
  word-break: break-all;
  height: 100%;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  .session-item {
    min-height: 80px;
    display: flex;
    // justify-content:flex-end;
    padding: 20px;
    line-height: 28px;
    img {
      width: 30px;
      height: 30px;
      object-fit: cover;
    }
    .logo {
      border-radius: 6px;
    }
    .answer-content {
      padding: 0 10px 10px 15px;
      position: relative;
      color: #333;
      .answer-content-query {
        display: flex;
        flex-wrap: wrap;
        flex-direction: column;
        align-items: flex-end;
        width: 100%;
        .answer-text {
          background: #7288fa;
          color: #fff;
          padding: 8px 10px 8px 20px;
          border-radius: 10px 0 10px 10px;
          margin: 0 !important;
          line-height: 1.5;
        }
        .session-setting-id {
          color: rgba(98, 98, 98, 0.5);
          font-size: 12px;
          margin-top: -8px;
        }
        .echo-doc-box {
          margin-top: 10px;
          width: 100%;
          max-width: 100%;
          display: flex;
          gap: 8px;
          justify-content: space-between;
          align-items: center;
          position: relative;
          .scroll-btn {
            position: absolute;
            top: 50%;
            transform: translateY(-15px);
            &.left {
              left: 5px;
            }
            &.right {
              right: 5px;
            }
          }
          .imgList {
            width: 100%;
            gap: 10px;
            overflow-x: hidden;
            scroll-behavior: smooth;
            display: flex;
            flex-wrap: nowrap;
            flex-direction: row-reverse;
          }
          .docInfo-container {
            display: flex;
            align-items: center;
            background: #fff;
            border: 1px solid rgb(235, 236, 238);
            padding: 5px 10px 5px 5px;
            border-radius: 5px;
          }
          .docInfo-img-container {
            flex-shrink: 0; /* 防止图片被压缩 */
            width: auto; /* 或固定宽度 */
            p {
              text-align: center;
              color: $color;
              font-size: 12px;
            }
          }
          .docIcon {
            width: 30px;
            height: 30px;
          }
          .imgIcon {
            width: auto !important;
            height: 70px !important;
            display: block;
            border-radius: 6px;
          }
          .docInfo {
            margin-left: 5px;
            .docInfo_name {
              color: #333;
            }
            .docInfo_size {
              color: #bbbbbb;
              text-align: left !important;
            }
          }
        }
      }
      li {
        display: revert !important;
      }
    }
  }
  .session-answer {
    border-radius: 10px;
    .answer-annotation {
      line-height: 0 !important;
      .annotation-img {
        width: 460px;
        object-fit: contain;
        height: 358px;
      }
      .tagging-canvas {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        margin: auto;
      }
    }

    .no-response {
      margin: 15px 0;
    }
    /*出处*/
    .search-list {
      padding: 10px 20px 3px 54px;
      .qa_content {
        display: flex;
        gap: 10px;
        margin-top: 5px;
      }
      .recommended-question-title {
        border-bottom: 1px solid #e5e5e5;
        padding: 5px 0;
      }
      .search-list-item {
        margin-bottom: 5px;
        line-height: 22px;
        p:nth-child(1) {
          white-space: normal;
        }
        a,
        span {
          color: #666;
          cursor: pointer;
          white-space: normal;
          overflow-wrap: break-word;
        }
        a {
          text-decoration: underline;
        }
        a:hover {
          color: deepskyblue;
        }
        .snippet {
          padding: 5px 14px;
        }
      }
    }
    /*操作*/
    .answer-operation {
      display: flex;
      // justify-content: space-between;
      align-items: center;
      padding: 5px 20px 15px 63px;
      color: #777;
      .opera-left {
        // flex: 8;
        .restart,
        .preStop {
          cursor: pointer;
          img {
            width: 20px;
            height: 20px;
            padding: 2px;
          }
        }
      }
      .opera-right {
        // flex: 1;
        cursor: pointer;
        display: inline-flex;
        padding-left: 10px;
        img {
          width: 20px;
          height: 20px;
          padding: 2px;
        }
        .split-icon {
          background: rgba(195, 197, 217, 0.65);
          height: 22px;
          margin: 0 10px;
          width: 1px;
        }
        .copy-icon {
          font-size: 17px;
          padding: 3px 6px;
          margin: 0 15px;
          cursor: pointer;
        }
        .copy-icon:hover {
          color: #33a4df;
        }
      }
      .answer-operation-tip {
        padding: 0 0 4px 10px;
        font-size: 12px;
        color: #999;
      }
    }
  }

  /*图片*/
  .file-path {
    .el-image {
      height: 200px !important;
      background-color: #f9f9f9;
      ::v-deep.el-image__inner,
      img {
        width: 100%;
        height: 100%;
        object-fit: contain;
      }
    }
    audio {
      width: 300px !important;
    }
  }
  .query-file {
    padding: 10px 0;
  }
  .response-file {
    margin: 0 0 0 66px;
    width: 400px;
    font-size: 0;
    .img {
      display: inline-block;
      width: 200px;
      height: 200px;
      img {
        width: 100%;
        height: 100%;
      }
    }
  }

  .session-error {
    background-color: #fef0f0;
    border-color: #fde2e2;
    color: #f56c6c !important;
    margin-top: 10px;
    padding: 10px;
    border-radius: 4px;
    .el-icon-warning {
      font-size: 16px;
    }
  }

  .history-box {
    height: calc(100% - 46px);
    flex: 1;
    overflow-y: auto !important;
    padding: 20px;
  }
  /*删除历史...*/
  .session-setting {
    position: relative;
    height: 36px;
    right: 50px;
    .right-setting {
      position: absolute;
      right: 10px;
      top: -5px;
      color: #ff2324;
      font-size: 16px;
      cursor: pointer;
      ::v-deep {
        .el-dropdown-menu {
          width: 100px;
        }
        .el-dropdown-menu__item {
          padding: 0 15px !important;
        }
      }
    }
  }

  .think_icon {
    width: 12px !important;
    height: 12px !important;
    margin-right: 3px;
  }
  .ds-res {
    ::v-deep section {
      color: #8b8b8b;
      position: relative;
      font-size: 12px;
      * {
        font-size: 12px;
      }
    }
    ::v-deep section::before {
      content: '';
      position: absolute;
      height: 100%;
      width: 1px;
      background: #ddd;
      left: -8px;
    }
    ::v-deep .hideDs {
      display: none;
    }
  }

  .deepseek {
    font-size: 13px;
    color: #8b8b8b;
    font-weight: bold;
    margin: 0 0 10px 6px;
    cursor: pointer;
    display: inline-block;
  }

  .sub-conversion-box {
    border-radius: 8px;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
}
/* 仅通过样式调整位置：
   问题在右侧（内容在右、头像在最右），答案在左侧（默认） */
.session-question {
  .session-item {
    flex-direction: row-reverse;
    margin-left: auto;
    width: auto;
  }
}
.session-answer {
  .session-answer-wrapper {
    display: flex;
    align-items: flex-start;
    gap: 10px; /* 头像和内容之间10px距离 */
    padding: 20px 20px 0 20px;
    min-height: 80px;
    background: none; /* 确保外层容器无背景色 */

    .logo {
      width: 30px;
      height: 30px;
      border-radius: 6px;
      object-fit: cover;
      flex-shrink: 0; /* 防止头像被压缩 */
      background: none; /* 头像无背景色 */
    }

    .answer-content {
      flex: 1;
      background-color: #eceefe; /* 只有内容区域有背景色 */
      border-radius: 0 10px 10px 10px;
      padding: 20px;
      line-height: 1.6;
    }
  }
}

/* 图片加载失败时的样式 */
img.failed {
  position: relative;
  border: 2px dashed #ff6b6b;
  background-color: #fff5f5;
  opacity: 0.5;
}

img.failed::after {
  content: '图片加载失败';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: #ff6b6b;
  font-size: 12px;
  background: rgba(255, 255, 255, 0.9);
  padding: 4px 8px;
  border-radius: 4px;
  white-space: nowrap;
}

.text-loading,
.text-loading > div {
  position: relative;
  box-sizing: border-box;
}

.text-loading {
  display: block;
  font-size: 0;
  color: #c8c8c8;
}

.text-loading.la-dark {
  color: #e8e8e8;
}

.text-loading > div {
  display: inline-block;
  float: none;
  background-color: currentColor;
  border: 0 solid currentColor;
}

.text-loading {
  width: 54px;
  height: 18px;
  margin: 6px 0 0 55px;
}

.text-loading > div {
  width: 8px;
  height: 8px;
  margin: 4px;
  border-radius: 100%;
  animation: ball-beat 0.7s -0.15s infinite linear;
}

.text-loading > div:nth-child(2n-1) {
  animation-delay: -0.5s;
}
@keyframes ball-beat {
  50% {
    opacity: 0.2;
    transform: scale(0.75);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

.session-section-wrapper {
  padding: 5px 20px 15px 63px;
}

.recommend-question {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
  &-item {
    display: inline-flex;
    font-size: 12px;
    padding: 4px 6px;
    background: #f2f2f2;
    cursor: pointer;
    border-radius: 8px;
    align-items: center;
    line-height: 14px;
    &:hover {
      background: #eceefe;
    }
    &.is-tips {
      cursor: default;
      opacity: 0.7;
      &:hover {
        background: #f2f2f2;
      }
    }
  }
  &-loading {
    margin: 0;
  }
}

.message-sequence-wrapper {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
</style>
