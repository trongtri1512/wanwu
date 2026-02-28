<template>
  <el-upload
    :class="{ hide: hideUpload }"
    action=""
    :multiple="false"
    list-type="picture-card"
    :accept="acceptType"
    :limit="1"
    :auto-upload="false"
    :on-change="uploadOnChange"
    :on-remove="handleRemove"
  >
    <i class="el-icon-picture-outline"></i>
  </el-upload>
</template>

<script>
import { uploadFile } from '@/api/chunkFile';

export default {
  model: {
    prop: 'value',
    event: 'input',
  },
  props: {
    value: {
      type: Object,
      default: {},
    },
    acceptType: {
      type: String,
      default: '',
    },
    maxSize: {
      type: Number,
      default: 3,
    },
  },
  data() {
    return {
      hideUpload: false,
    };
  },
  methods: {
    uploadOnChange(file) {
      if (file) {
        if (file.size / 1024 / 1024 > this.maxSize) {
          this.$message.error(
            this.$t('knowledgeManage.multiKnowledgeDatabase.imageSizeLimit', {
              maxSize: this.maxSize,
            }),
          );
          return;
        }
        this.hideUpload = true;
        const formData = new FormData();
        const config = { headers: { 'Content-Type': 'multipart/form-data' } };
        formData.append('files', file.raw);
        uploadFile(formData, config).then(res => {
          if (res.code === 0) {
            this.$emit('input', res.data.files[0]);
          }
        });
      }
    },
    handleRemove() {
      this.hideUpload = false;
      this.$emit('input', {});
    },
  },
};
</script>

<style lang="scss" scoped>
::v-deep {
  .el-upload--picture-card,
  .el-upload-list__item {
    width: 30px;
    height: 30px;
    display: flex;
    justify-content: center;
    text-align: center;
  }
}

.hide ::v-deep .el-upload--picture-card {
  display: none;
}
</style>
