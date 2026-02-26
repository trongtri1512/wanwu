<template>
  <div class="createDialog">
    <el-dialog
      :visible.sync="dialogVisible"
      width="760px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
    >
      <template slot="title">
        <div class="dialog-title-wrapper">
          <span class="dialog-title">{{ provider.name || '' }}</span>
          <span class="dialog-desc" v-if="provider.key === yuanjing">
            {{ $t('modelAccess.hint.yuanjing') }}
          </span>
        </div>
      </template>
      <el-form
        :model="{ ...createForm }"
        :rules="rules"
        ref="createForm"
        label-width="130px"
        class="createForm form"
      >
        <el-form-item
          :label="$t('modelAccess.table.modelType')"
          prop="modelType"
        >
          <el-select
            v-model="createForm.modelType"
            :placeholder="$t('common.select.placeholder')"
            :disabled="isEdit"
            style="width: 100%"
          >
            <el-option
              v-for="item in modelType"
              :key="item.key"
              :label="item.name"
              :value="item.key"
            ></el-option>
          </el-select>
          <div
            v-if="
              createForm.modelType === embedding && provider.key === yuanjing
            "
            class="embedding-tip"
          >
            {{ $t('modelAccess.table.embeddingTip') }}
          </div>
        </el-form-item>
        <el-form-item :label="$t('modelAccess.table.modelName')" prop="model">
          <el-input
            :disabled="isEdit"
            v-model="createForm.model"
            :placeholder="$t('common.input.placeholder')"
          ></el-input>
        </el-form-item>
        <el-form-item
          :label="$t('modelAccess.table.modelDisplayName')"
          prop="displayName"
        >
          <el-input
            v-model="createForm.displayName"
            :placeholder="$t('common.hint.modelName')"
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('modelAccess.table.picPath')" prop="avatar">
          <el-upload
            class="avatar-uploader"
            action=""
            name="files"
            :show-file-list="false"
            :http-request="handleUploadImage"
            :on-error="handleUploadError"
            accept=".png,.jpg,.jpeg"
          >
            <img
              class="upload-img"
              :src="
                createForm.avatar && createForm.avatar.path
                  ? avatarSrc(createForm.avatar.path)
                  : defaultLogo
              "
            />
            <!--<span style="margin-left: 12px; color: #606266 !important;" v-if="createForm.avatar && createForm.avatar.path">
              {{createForm.avatar.path}}
            </span>-->
            <span class="upload-hint">{{ $t('modelAccess.hint.upload') }}</span>
          </el-upload>
        </el-form-item>
        <el-form-item
          :label="$t('modelAccess.table.modelDesc')"
          prop="modelDesc"
        >
          <el-input
            type="text"
            v-model="createForm.modelDesc"
            :placeholder="$t('common.input.placeholder')"
          ></el-input>
        </el-form-item>
        <el-form-item
          v-if="createForm.modelType === llm"
          label="Function Call"
          prop="functionCalling"
        >
          <el-select
            v-model="createForm.functionCalling"
            :placeholder="$t('common.select.placeholder')"
            style="width: 100%"
          >
            <el-option
              v-for="item in functionCalling"
              :key="item.key"
              :label="item.name"
              :value="item.key"
            ></el-option>
          </el-select>
        </el-form-item>
        <el-form-item v-if="showVision()" label="Vision" prop="visionSupport">
          <el-select
            v-model="createForm.visionSupport"
            :placeholder="$t('common.select.placeholder')"
            style="width: 100%"
          >
            <el-option
              v-for="item in supportList"
              :key="item.key"
              :label="item.name"
              :value="item.key"
            ></el-option>
          </el-select>
        </el-form-item>
        <!-- 多模态模型去掉支持的文件类型选择 -->
        <!--<el-form-item
          v-if="isMultiModal()"
          :label="$t('modelAccess.table.supportFileType')"
          prop="supportFileTypes"
        >
          <el-select
            v-model="createForm.supportFileTypes"
            :placeholder="$t('common.select.placeholder')"
            style="width: 100%"
            multiple
          >
            <el-option
              v-for="item in Object.keys(supportFileTypeObj)"
              :key="item"
              :label="supportFileTypeObj[item]"
              :value="item"
            ></el-option>
          </el-select>
        </el-form-item>-->
        <el-form-item
          v-if="showMaxPicLimit()"
          :label="$t('modelAccess.table.maxPicLimit')"
          prop="maxImageSize"
        >
          <el-input-number
            v-model="createForm.maxImageSize"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
          ></el-input-number>
          M
        </el-form-item>
        <!-- 多模态模型去掉视频片限制，和最大文本长度 -->
        <!--<el-form-item
          v-if="showMaxVideoLimit()"
          :label="$t('modelAccess.table.maxVideoLimit')"
          prop="maxVideoClipSize"
        >
          <el-input-number
            v-model="createForm.maxVideoClipSize"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
          ></el-input-number>
          M
        </el-form-item>
        <el-form-item
          v-if="isMultiModal()"
          :label="$t('modelAccess.table.maxTextSize')"
          prop="maxTextLength"
        >
          <el-input-number
            v-model="createForm.maxTextLength"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
          ></el-input-number>
          tokens
        </el-form-item>-->
        <el-form-item
          v-if="showMaxAudioLimit()"
          :label="$t('modelAccess.table.maxAudioLimit')"
          prop="maxAsrFileSize"
        >
          <el-input-number
            v-model="createForm.maxAsrFileSize"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
          ></el-input-number>
          M
        </el-form-item>
        <el-form-item
          v-if="showContextSize()"
          :label="$t('modelAccess.table.contextSize')"
          prop="contextSize"
        >
          <el-input-number
            v-model="createForm.contextSize"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
          ></el-input-number>
          tokens
        </el-form-item>
        <el-form-item
          v-if="createForm.modelType === llm"
          label="Max_token"
          prop="maxTokens"
        >
          <el-input-number
            v-model="createForm.maxTokens"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
          ></el-input-number>
          tokens
        </el-form-item>
        <el-form-item
          v-if="provider.key !== ollama && !showAppAndAccessKey()"
          :label="$t('modelAccess.table.apiKey')"
          prop="apiKey"
        >
          <el-input
            type="password"
            v-model="createForm.apiKey"
            :placeholder="
              $t('common.hint.apiKey') + (typeObj.apiKey[provider.key] || '--')
            "
          ></el-input>
        </el-form-item>
        <div v-if="showAppAndAccessKey()">
          <el-form-item label="APP Key" prop="appKey">
            <el-input
              type="password"
              v-model="createForm.appKey"
              :placeholder="$t('common.hint.appKey')"
            ></el-input>
          </el-form-item>
          <el-form-item label="Access Key" prop="accessKey">
            <el-input
              type="password"
              v-model="createForm.accessKey"
              :placeholder="$t('common.hint.accessKey')"
            ></el-input>
          </el-form-item>
        </div>
        <el-form-item
          :label="$t('modelAccess.table.inferUrl')"
          prop="endpointUrl"
        >
          <el-input
            v-model="createForm.endpointUrl"
            :title="
              $t('common.hint.inferUrl') +
              (typeObj.inferUrl[`${createForm.modelType}_${provider.key}`] ||
                typeObj.inferUrl[provider.key] ||
                '--')
            "
            :placeholder="
              $t('common.hint.inferUrl') +
              (typeObj.inferUrl[`${createForm.modelType}_${provider.key}`] ||
                typeObj.inferUrl[provider.key] ||
                '--')
            "
          ></el-input>
        </el-form-item>
        <!--<el-form-item :label="$t('modelAccess.table.publishTime')" prop="publishDate">
          <el-date-picker
            v-model="createForm.publishDate"
            type="date"
            value-format="yyyy-MM-dd"
            :placeholder="$t('common.select.placeholder')"
          >
          </el-date-picker>
        </el-form-item>-->
        <el-form-item v-if="isEdit" label="uuid">
          {{ row.uuid || '--' }}
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="handleClose">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button :loading="loading" type="primary" @click="handleSubmit">
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>
<script>
import { addModel, editModel } from '@/api/modelAccess';
import { uploadAvatar } from '@/api/user';
import { avatarSrc } from '@/utils/util';
import {
  PROVIDER_TYPE,
  PROVIDER_OBJ,
  FUNC_CALLING,
  DEFAULT_CALLING,
  DEFAULT_SUPPORT,
  SUPPORT_LIST,
  TYPE_OBJ,
  LLM,
  RERANK,
  EMBEDDING,
  MULTIMODAL_EMBEDDING,
  MULTIMODAL_RERANK,
  ASR,
  OLLAMA,
  YUAN_JING,
  QWEN,
  QIANFAN,
  SUPPORT_FILE_TYPE_OBJ,
  IMAGE,
  VIDEO,
  HUOSHAN,
} from '../constants';
import LinkIcon from '@/components/linkIcon.vue';

export default {
  components: { LinkIcon },
  data() {
    const validateUrls = (rule, value, callback) => {
      const reg =
        /^(http|ftp|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&:/~\+#]*[\w\-\@?^=%&/~\+#])?$/;

      if (!reg.test(value)) {
        callback(new Error(this.$t('modelAccess.hint.urlError')));
      } else {
        return callback();
      }
    };
    return {
      basePath: this.$basePath,
      defaultLogo: require('@/assets/imgs/model_default_icon.png'),
      dialogVisible: false,
      modelType: [],
      functionCalling: FUNC_CALLING,
      supportList: SUPPORT_LIST,
      supportFileTypeObj: SUPPORT_FILE_TYPE_OBJ,
      typeObj: TYPE_OBJ,
      llm: LLM,
      embedding: EMBEDDING,
      ollama: OLLAMA,
      yuanjing: YUAN_JING,
      showVisionList: [YUAN_JING, QWEN, QIANFAN],
      showContextSizeList: [
        LLM,
        EMBEDDING,
        RERANK,
        MULTIMODAL_RERANK,
        MULTIMODAL_EMBEDDING,
        ASR,
      ],
      createForm: {
        model: '',
        displayName: '',
        endpointUrl: '',
        apiKey: '',
        appKey: '',
        accessKey: '',
        modelType: LLM,
        modelDesc: '',
        contextSize: 8000,
        maxTokens: 4096,
        maxAsrFileSize: 10,
        maxImageSize: 3,
        /*maxVideoClipSize: 10,
        maxTextLength: 512,
        supportFileTypes: [IMAGE, VIDEO],*/
        avatar: {
          key: '',
          path: '',
        },
        // publishDate: '',
        functionCalling: DEFAULT_CALLING,
        visionSupport: DEFAULT_SUPPORT,
      },
      rules: {
        model: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          // { min: 2, max: 50, message: this.$t('common.hint.modelNameLimit'), trigger: 'blur'},
          // { pattern: /^(?!_)[a-zA-Z0-9-_.\u4e00-\u9fa5]+$/, message: this.$t('common.hint.modelName'), trigger: "blur"}
        ],
        appKey: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        accessKey: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        contextSize: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        maxTokens: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        displayName: [
          {
            pattern: /^(?!_)[a-zA-Z0-9-_.\u4e00-\u9fa5]+$/,
            message: this.$t('common.hint.modelName'),
            trigger: 'blur',
          },
          {
            min: 2,
            max: 50,
            message: this.$t('common.hint.modelNameLimit'),
            trigger: 'blur',
          },
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        modelType: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
        endpointUrl: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          { validator: validateUrls, trigger: 'blur' },
        ],
        /*supportFileTypes: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],*/
      },
      row: {},
      provider: {},
      isEdit: false,
      loading: false,
    };
  },
  watch: {
    'createForm.modelType': {
      handler() {
        this.$refs.createForm.clearValidate();
        if (!this.isEdit) {
          this.setDefaultInferUrl();
        }
      },
      immediate: false,
    },
    'provider.key': {
      handler(newVal) {
        if (!this.isEdit && newVal) {
          this.setDefaultInferUrl();
        }
      },
      immediate: false,
    },
  },
  methods: {
    avatarSrc,
    isMultiModal() {
      return [MULTIMODAL_RERANK, MULTIMODAL_EMBEDDING].includes(
        this.createForm.modelType,
      );
    },
    showAppAndAccessKey() {
      return this.provider.key === HUOSHAN && this.createForm.modelType === ASR;
    },
    showVision() {
      return (
        this.createForm.modelType === LLM &&
        this.showVisionList.includes(this.provider.key)
      );
    },
    showContextSize() {
      return this.showContextSizeList.includes(this.createForm.modelType);
    },
    showFileTypeLimit(type) {
      return (
        this.isMultiModal() && this.createForm.supportFileTypes.includes(type)
      );
    },
    showMaxPicLimit() {
      const { modelType, visionSupport } = this.createForm || {};
      return (
        (modelType === LLM && visionSupport === 'support') ||
        this.isMultiModal() // this.showFileTypeLimit(IMAGE)
      );
    },
    showMaxVideoLimit() {
      return this.showFileTypeLimit(VIDEO);
    },
    showMaxAudioLimit() {
      return this.createForm.modelType === ASR;
    },
    uploadAvatar(file, key) {
      const formData = new FormData();
      const config = { headers: { 'Content-Type': 'multipart/form-data' } };
      formData.append(key, file);
      return uploadAvatar(formData, config);
    },
    handleUploadImage(data) {
      if (data.file) {
        this.uploadAvatar(data.file, 'avatar').then(res => {
          const { key, path } = res.data || {};
          this.createForm.avatar = { key, path };
        });
      }
    },
    handleUploadError() {
      this.$message.error(this.$t('common.message.uploadError'));
    },
    formatValue(data) {
      for (let key in this.createForm) {
        this.createForm[key] = data ? data[key] || '' : '';
      }
    },
    setDefaultInferUrl() {
      const defaultUrl =
        this.typeObj.inferUrl[
          `${this.createForm.modelType}_${this.provider.key}`
        ] || this.typeObj.inferUrl[this.provider.key];
      if (defaultUrl) {
        this.createForm.endpointUrl = defaultUrl;
      }
    },
    openDialog(title, row) {
      this.provider = { key: title, name: PROVIDER_OBJ[title] };
      const currentProvider =
        PROVIDER_TYPE.find(item => item.key === title) || {};
      this.modelType = currentProvider.children || [];
      this.createForm.modelType = this.modelType[0]
        ? this.modelType[0].key || LLM
        : LLM;

      // 自动填入推理URL默认值
      if (!row) {
        this.setDefaultInferUrl();
      }

      this.dialogVisible = true;

      this.isEdit = Boolean(row);
      if (this.isEdit) {
        this.row = row || {};
        this.formatValue(row);
      }
    },
    handleClose() {
      this.dialogVisible = false;
      this.formatValue({
        modelType: LLM,
        functionCalling: DEFAULT_CALLING,
        visionSupport: DEFAULT_SUPPORT,
        contextSize: 8000,
        maxTokens: 4096,
        maxAsrFileSize: 10,
        maxImageSize: 3,
        /*maxVideoClipSize: 10,
        maxTextLength: 512,
        supportFileTypes: [IMAGE, VIDEO],*/
        avatar: { key: '', path: '' },
      });
      if (this.$refs.createForm) {
        this.$refs.createForm.resetFields();
        this.$refs.createForm.clearValidate();
      }
    },
    handleSubmit() {
      this.$refs.createForm.validate(async valid => {
        if (valid) {
          const {
            apiKey,
            appKey,
            accessKey,
            endpointUrl,
            functionCalling,
            modelType,
            visionSupport,
            contextSize,
            maxTokens,
            /*maxTextLength,
            maxVideoClipSize,
            supportFileTypes,*/
            maxImageSize,
            maxAsrFileSize,
          } = this.createForm;
          const form = {
            ...this.createForm,
            provider: this.provider.key || '',
            config: {
              endpointUrl,
              ...(this.provider.key !== OLLAMA &&
                !this.showAppAndAccessKey() && { apiKey }),
              ...(modelType === LLM && { functionCalling, maxTokens }),
              ...(this.showVision() && { visionSupport }),
              ...(this.showContextSize() && { contextSize }),
              ...(this.showMaxAudioLimit() && { maxAsrFileSize }),
              ...(this.showMaxPicLimit() && { maxImageSize }),
              ...(this.showAppAndAccessKey() && { appKey, accessKey }),
              /*...(this.showMaxVideoLimit() && { maxVideoClipSize }),
              ...(this.isMultiModal() && { supportFileTypes, maxTextLength }),*/
            },
          };
          const deleteKeys = [
            'apiKey',
            'appKey',
            'accessKey',
            'endpointUrl',
            'functionCalling',
            'visionSupport',
            'contextSize',
            'maxTokens',
            /*'maxTextLength',
            'maxVideoClipSize',
            'supportFileTypes',*/
            'maxImageSize',
            'maxAsrFileSize',
          ];
          deleteKeys.forEach(key => {
            delete form[key];
          });

          try {
            this.loading = true;
            const res = this.isEdit
              ? await editModel({ ...form, modelId: this.row.modelId })
              : await addModel(form);
            if (res.code === 0) {
              this.$message.success(this.$t('common.message.success'));
              this.handleClose();
              this.$emit('reloadData', !this.isEdit);
            }
          } finally {
            this.loading = false;
          }
        }
      });
    },
  },
};
</script>
<style lang="scss" scoped>
.createForm {
  padding: 0 45px 0 20px;
  .avatar-uploader {
    .upload-img {
      object-fit: cover;
      width: 80px;
      height: 80px;
      border-radius: 8px;
      border: 1px solid #dcdfe6;
      display: inline-block;
      vertical-align: middle;
    }
    .upload-hint {
      display: inline-block;
      margin-left: 12px;
      color: #909399 !important;
    }
  }
  .embedding-tip {
    color: #f56c6c;
    line-height: 16px;
  }
}
.dialog-title-wrapper {
  display: flex;
  align-items: center;
  .dialog-title {
    color: $color_title;
    font-size: 18px;
    font-weight: bold;
  }
  .dialog-desc {
    color: #888;
    margin-left: 20px;
  }
}
</style>
