<template>
  <el-upload
    class="avatar-uploader"
    action=""
    name="files"
    :show-file-list="false"
    :multiple="false"
    :http-request="handleUploadAvatar"
    :on-error="handleUploadError"
    accept=".png,.jpg,.jpeg"
  >
    <img class="upload-img" :src="avatarSrc" alt="" />
    <p class="upload-hint">
      {{ this.$t('common.fileUpload.clickUploadImg') }}
    </p>
  </el-upload>
</template>

<script>
import { uploadAvatar } from '@/api/user';
import { avatarSrc } from '@/utils/util';

export default {
  model: {
    prop: 'value',
    event: 'input',
  },
  props: {
    value: {
      type: Object,
      default: () => ({
        key: '',
        path: '',
      }),
    },
    // 默认头像，导入需要require("@/assets/imgs/defaultAvatar")
    defaultAvatar: {
      type: String,
      default: '',
    },
  },
  computed: {
    avatarSrc() {
      if (this.value.path) {
        return avatarSrc(this.value.path);
      }
      return this.defaultAvatar;
    },
  },
  methods: {
    handleUploadAvatar(data) {
      if (data.file) {
        const formData = new FormData();
        const config = { headers: { 'Content-Type': 'multipart/form-data' } };
        formData.append('avatar', data.file);
        uploadAvatar(formData, config).then(res => {
          if (res.code === 0) {
            this.$emit('input', res.data);
          }
        });
      }
    },
    handleUploadError() {
      this.$message.error(this.$t('common.message.uploadError'));
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
    width: 98px;
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
  }
}
</style>
