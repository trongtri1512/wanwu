<template>
  <div
    class="rich-textarea-wrapper"
    :class="{
      'read-only': isReadOnly,
    }"
  >
    <div
      class="rich-textarea"
      ref="editorRef"
      :contenteditable="!isReadOnly"
      @compositionstart="onCompositionStart"
      @compositionend="onCompositionEnd"
      @focus="onFocus"
      @blur="onBlur"
      @input="onInput"
      v-html="Md2Img(currentValue)"
    ></div>

    <!-- 只在非只读模式下显示上传按钮 -->
    <el-upload
      v-if="!isReadOnly"
      action=""
      :multiple="false"
      :accept="acceptType"
      :auto-upload="false"
      :show-file-list="false"
      :on-change="uploadOnChange"
    >
      <i class="el-icon-picture-outline"></i>
    </el-upload>
  </div>
</template>

<script>
import { Md2Img, Img2Md } from '@/utils/util';
import { POWER_TYPE_READ } from '@/views/knowledge/constants';
import { getDocLimit } from '@/api/knowledge';
import { uploadFileMD } from '@/api/chunkFile';

export default {
  props: {
    value: {
      type: String,
      default: '',
    },
    permissionType: {
      type: Number,
      default: POWER_TYPE_READ,
    },
    knowledgeId: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      savedRange: null,
      isComposing: false, // 用于处理中文输入法等复杂输入场景
      acceptType: '.png,.jpg,.jpeg',
      maxSize: 3,
    };
  },
  computed: {
    isReadOnly() {
      return [POWER_TYPE_READ].includes(this.permissionType);
    },
    currentValue() {
      return this.value;
    },
  },
  watch: {
    knowledgeId: {
      handler(newVal) {
        if (newVal) {
          getDocLimit({ knowledgeId: newVal }).then(res => {
            if (res.code === 0) {
              this.acceptType =
                '.' +
                res.data.uploadLimitList
                  .find(item => item.acceptType === 'image')
                  .flatMap(item => item.extList || [])
                  .join(',.');
              this.maxSize = res.data.uploadLimitList.find(
                item => item.fileType === 'image',
              ).maxSize;
            }
          });
        }
      },
    },
    currentValue(newVal) {
      if (newVal !== this.localContent) {
        this.localContent = newVal;
      }
    },
  },
  created() {
    this.localContent = this.currentValue;
  },
  methods: {
    Md2Img,

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
        const formData = new FormData();
        const config = { headers: { 'Content-Type': 'multipart/form-data' } };
        formData.append('files', file.raw);
        formData.append('markdown', true);
        uploadFileMD(formData, config).then(res => {
          if (res.code === 0) {
            this.insertImage(res.data.fileList[0].fileUrl);
          }
        });
      }
    },

    onCompositionStart() {
      // 开始输入法输入，标记状态
      this.isComposing = true;
    },

    onCompositionEnd() {
      // 输入法输入结束
      if (this.isComposing) {
        this.isComposing = false;
      }
    },

    onFocus() {
      if (this.isReadOnly) return;
      this.saveCaretPosition();
    },

    onBlur() {
      if (this.isReadOnly) return;
      this.saveCaretPosition();
    },

    onInput(event) {
      if (this.isReadOnly) return;
      // 对于复杂输入（如中文输入法），我们不立即保存，而是等待输入完成
      if (this.isComposing) return;

      const currentHTML = event.target.innerHTML;
      this.localContent = currentHTML;
      this.$emit('input', currentHTML);
      // 保存光标位置，因为内容变化后光标位置可能改变
      this.saveCaretPosition();
    },

    saveCaretPosition() {
      const selection = window.getSelection();
      if (selection && selection.rangeCount > 0) {
        this.savedRange = selection.getRangeAt(0).cloneRange();
      }
    },

    insertImage(mdImageLink) {
      // 使用之前保存的光标位置，如果没有则添加到末尾
      if (this.savedRange) {
        const range = this.savedRange;

        // 删除当前选中内容
        range.deleteContents();

        // 创建一个临时节点包含 Markdown 链接
        const textNode = document.createTextNode(mdImageLink);
        range.insertNode(textNode);

        // 更新光标位置
        const newRange = document.createRange();
        newRange.setStartAfter(textNode);
        newRange.collapse(true);

        const selection = window.getSelection();
        selection.removeAllRanges();
        selection.addRange(newRange);

        this.savedRange = newRange.cloneRange();

        // 触发内容更新
        this.$nextTick(() => {
          const newHTML = this.$refs.editorRef.innerHTML;
          this.localContent = Img2Md(newHTML);
          this.$emit('input', this.localContent);
        });
      } else {
        // 如果没有保存的光标位置，直接添加到末尾
        const currentHTML = this.$refs.editorRef.innerHTML;
        this.localContent = Img2Md(currentHTML + mdImageLink);
        this.$emit('input', this.localContent);
      }
    },
  },
};
</script>

<style scoped>
.rich-textarea-wrapper {
  min-height: 40px;
  max-height: 150px;
  padding: 8px 12px;
  text-align: start;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  color: #606266;
  background-color: #fff;
  font-family: inherit;
  font-size: inherit;
  line-height: 1.5;
  outline: none;
  resize: vertical;
  overflow-y: auto;

  &.read-only {
    background-color: #f5f7fa;
    border-color: #e4e7ed;
    color: #c0c4cc;
    cursor: not-allowed;
  }

  ::v-deep img {
    width: auto;
    max-height: 115px;
  }

  .rich-textarea {
    outline: none;
  }

  &:hover,
  &:focus {
    border: 1px solid var(--color);
  }
}
</style>
