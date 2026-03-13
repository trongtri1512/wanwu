<template>
  <div>
    <el-dialog
      :title="titleMap[type]"
      :visible.sync="dialogVisible"
      width="750"
      append-to-body
      :close-on-click-modal="false"
      @closed="clearForm"
    >
      <el-form ref="form" :model="form" label-width="130px" :rules="rules">
        <el-form-item
          :label="$t('tempSquare.skills.form.avatar') + ':'"
          prop="avatar"
        >
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
                form.avatar && form.avatar.path
                  ? avatarSrc(form.avatar.path)
                  : defaultLogo
              "
            />
            <p class="upload-hint">
              {{ $t('common.fileUpload.clickUploadImg') }}
            </p>
          </el-upload>
        </el-form-item>
        <el-form-item
          :label="$t('tempSquare.skills.form.upload') + ':'"
          prop="zipUrl"
        >
          <div style="display: flex; align-items: center; gap: 10px">
            <el-upload
              ref="upload"
              action=""
              :show-file-list="false"
              :auto-upload="false"
              :on-change="uploadOnChange"
              accept=".zip"
              :limit="1"
            >
              <el-button size="mini" type="primary">
                {{ $t('tempSquare.skills.form.uploadPlaceholder') }}
              </el-button>
            </el-upload>
            <div class="el-upload__tip" style="margin-top: 0">
              {{ $t('tempSquare.skills.form.uploadTips') }}
            </div>
          </div>
          <!-- 上传文件预览列表 -->
          <div class="file-list-preview" v-if="fileList.length > 0">
            <div
              v-for="(file, index) in fileList"
              :key="index"
              class="file-item"
            >
              <div class="file-info">
                <i class="el-icon-document"></i>
                <span class="file-name">{{ file.name }}</span>
                <span class="file-size">({{ filterSize(file.size) }})</span>
                <div class="file-status-icon">
                  <i
                    v-if="file.progressStatus === 'success'"
                    class="el-icon-circle-check success-color"
                  ></i>
                  <i
                    v-else-if="file.progressStatus === 'exception'"
                    class="el-icon-circle-close fail-color"
                  ></i>
                  <i
                    v-else-if="file.percentage < 100"
                    class="el-icon-loading"
                  ></i>
                </div>
                <i
                  class="el-icon-delete delete-btn"
                  @click="handleRemove(file, index)"
                ></i>
              </div>
              <el-progress
                v-if="file.percentage < 100"
                :percentage="file.percentage"
                :stroke-width="2"
                :show-text="false"
              ></el-progress>
            </div>
          </div>
        </el-form-item>
        <el-form-item
          :label="$t('tempSquare.skills.form.author') + ':'"
          prop="author"
        >
          <el-input
            v-model="form.author"
            maxlength="30"
            show-word-limit
            :placeholder="$t('tempSquare.skills.form.authorPlaceholder')"
          ></el-input>
        </el-form-item>
        <el-form-item
          :label="$t('tempSquare.skills.form.name') + ':'"
          prop="name"
        >
          <el-input
            disabled
            v-model="form.name"
            maxlength="30"
            show-word-limit
          ></el-input>
        </el-form-item>
        <el-form-item
          :label="$t('tempSquare.skills.form.desc') + ':'"
          prop="desc"
        >
          <el-input
            type="textarea"
            disabled
            v-model="form.desc"
            show-word-limit
            maxlength="100"
          ></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button type="primary" @click="doSubmit" :loading="btnLoading">
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { uploadAvatar } from '@/api/user';
import { createCustomSkill, checkCustomSkill } from '@/api/templateSquare';
import { avatarSrc, filterSize } from '@/utils/util';
import uploadChunk from '@/mixins/uploadChunk';

export default {
  name: 'SkillDialog',
  mixins: [uploadChunk],
  props: {
    type: {
      type: String,
      default: 'import',
    },
  },
  data() {
    return {
      dialogVisible: false,
      btnLoading: false,
      defaultLogo: require('@/assets/imgs/custom-skill-default-icon.png'),
      form: {
        name: '',
        desc: '',
        author: '',
        avatar: {
          key: '',
          path: '',
        },
        zipUrl: '',
      },
      titleMap: {
        import: this.$t('tempSquare.skills.form.importTitle'),
      },
      skillId: '',
      rules: {
        avatar: [
          {
            required: false,
            message: this.$t('tempSquare.skills.formRules.avatar'),
            trigger: 'change',
          },
        ],
        name: [
          {
            required: false,
            message: this.$t('tempSquare.skills.formRules.name'),
            trigger: 'blur',
          },
        ],
        desc: [
          {
            required: false,
            message: this.$t('tempSquare.skills.formRules.desc'),
            trigger: 'blur',
          },
        ],
        zipUrl: [
          {
            required: true,
            message: this.$t('tempSquare.skills.formRules.zipUrl'),
            trigger: 'change',
          },
        ],
      },
    };
  },
  methods: {
    avatarSrc,
    filterSize,
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
          this.form.avatar = { key, path };
        });
      }
    },
    handleUploadError() {
      this.$message.error(this.$t('common.message.uploadError'));
    },
    uploadOnChange(file) {
      if (!this.validateZipFile(file.raw)) {
        this.$refs.upload.clearFiles();
        return;
      }

      this.fileList = [file];
      this.startUpload();
    },
    validateZipFile(file) {
      const isZip = file.name.endsWith('.zip');
      const isLt20M = file.size / 1024 / 1024 < 20;

      if (!isZip) {
        this.$message.error(this.$t('tempSquare.skills.formRules.zipFormat'));
        return false;
      }
      if (!isLt20M) {
        this.$message.error(this.$t('tempSquare.skills.formRules.zipSize'));
        return false;
      }
      return true;
    },
    // 覆盖 mixin 中的 uploadFile
    async uploadFile(fileName, originalName, filePath) {
      this.form.zipUrl = filePath;
      this.$refs.form.validateField('zipUrl');
      await this.checkSkillZip(filePath);
    },
    async checkSkillZip(filePath) {
      try {
        const res = await checkCustomSkill({ zipUrl: filePath });
        if (res.code === 0 && res.data) {
          this.form.name = res.data.name || '';
          this.form.desc = res.data.desc || '';
          this.$refs.form.validateField(['name', 'desc']);
        } else {
          if (this.fileList && this.fileList.length > 0) {
            this.handleRemove(this.fileList[0], 0);
          }
        }
      } catch (error) {
        if (this.fileList && this.fileList.length > 0) {
          this.handleRemove(this.fileList[0], 0);
        }
      }
    },
    handleRemove(file, index) {
      this.fileList.splice(index, 1);
      this.form.zipUrl = '';
      this.form.name = '';
      this.form.desc = '';
      this.$nextTick(() => {
        this.$refs.upload.clearFiles();
      });
    },
    openDialog(row) {
      if (row) {
        const { skillId, name, desc, avatar, prompt, zipUrl } = row;
        this.skillId = skillId;
        this.form = {
          name,
          desc,
          avatar: avatar || { key: '', path: '' },
          prompt,
          zipUrl,
        };
      } else {
        this.clearForm();
      }
      this.dialogVisible = true;
      this.$nextTick(() => {
        this.$refs['form'].clearValidate();
      });
    },
    clearForm() {
      this.form = {
        name: '',
        desc: '',
        author: '',
        avatar: {
          key: '',
          path: '',
        },
        zipUrl: '',
      };
      this.fileList = [];
      this.skillId = '';
      this.$nextTick(() => {
        this.$refs.upload && this.$refs.upload.clearFiles();
      });
    },
    async doSubmit() {
      await this.$refs.form.validate(async valid => {
        if (valid) {
          this.btnLoading = true;
          try {
            const { author, avatar, zipUrl } = this.form;
            const res = await createCustomSkill({ author, avatar, zipUrl });
            if (res.code === 0) {
              this.$message.success(this.$t('common.message.success'));
              this.dialogVisible = false;
              this.$emit('reload');
            }
          } catch (error) {
            console.error(error);
          } finally {
            this.btnLoading = false;
          }
        }
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.avatar-uploader {
  position: relative;
  width: 98px;
  .upload-img {
    object-fit: cover;
    width: 100%;
    height: 98px;
    background: #eee;
    border-radius: 8px;
    border: 1px solid #dcdfe6;
    display: inline-block;
    vertical-align: middle;
  }
  .upload-hint {
    position: absolute;
    width: 100%;
    bottom: 0;
    background: $color_opacity;
    color: $color;
    font-size: 12px;
    line-height: 26px;
    z-index: 10;
    border-radius: 0 0 8px 8px;
    text-align: center;
  }
}
.file-list-preview {
  margin-top: 10px;
  .file-item {
    background: #f5f7fa;
    padding: 8px 12px;
    border-radius: 4px;
    margin-bottom: 5px;
    .file-info {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 13px;
      .file-name {
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .file-size {
        color: #909399;
      }
      .file-status-icon {
        width: 16px;
        .success-color {
          color: #67c23a;
        }
        .fail-color {
          color: #f56c6c;
        }
      }
      .delete-btn {
        cursor: pointer;
        color: #f56c6c;
        &:hover {
          opacity: 0.8;
        }
      }
    }
    ::v-deep .el-progress {
      margin-top: 5px;
    }
  }
}
</style>
