<template>
  <el-select
    v-model="currentValue"
    :placeholder="placeholder"
    value-key="modelId"
    :disabled="disabled"
    @change="handleChange"
    @visible-change="handleVisibleChange"
    :loading-text="loadingText"
    :loading="loading"
    :filterable="filterable"
    :clearable="clearable"
    :popper-class="popperClass"
    class="model-select"
    ref="modelSelect"
  >
    <template v-if="visibleModelAvatar" #prefix>
      <img class="model-img" :src="modelAvatar" />
    </template>
    <el-option
      v-for="item in options"
      :key="item.modelId"
      :label="item.displayName"
      :value="item.modelId"
      @click.native="handleOptionClick(item)"
    >
      <div class="model-option-content">
        <span class="model-name">{{ item.displayName }}</span>
        <div class="model-select-tags" v-if="item.tags && item.tags.length > 0">
          <span
            v-for="(tag, tagIdx) in item.tags"
            :key="tagIdx"
            class="model-select-tag"
          >
            {{ tag.text }}
          </span>
        </div>
      </div>
    </el-option>
  </el-select>
</template>

<script>
import { avatarSrc } from '@/utils/util';
import defaultModelAvatar from '@/assets/imgs/model_default_icon.png';
export default {
  name: 'ModelSelect',
  props: {
    value: {
      type: [String, Number],
      default: '',
    },
    options: {
      type: Array,
      default: () => [],
    },
    placeholder: {
      type: String,
      default: function () {
        return this.$t('common.select.placeholder');
      },
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    loadingText: {
      type: String,
      default: function () {
        return this.$t('tempSquare.loadingText');
      },
    },
    loading: {
      type: Boolean,
      default: false,
    },
    filterable: {
      type: Boolean,
      default: false,
    },
    clearable: {
      type: Boolean,
      default: false,
    },
    popperClass: {
      type: String,
      default: '',
    },
    warning: {
      type: Boolean,
      default: false,
    },
    // 显示模型头像
    visibleModelAvatar: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      currentValue: this.value,
    };
  },
  watch: {
    value(val) {
      this.currentValue = val;
    },
    currentValue(val) {
      this.$emit('input', val);
    },
    loading(val) {
      // 加载完成之后，重新计算下拉框的位置以解决下拉框位置计算错误
      if (!val) {
        this.$nextTick(() => {
          if (this.$refs.modelSelect && this.$refs.modelSelect.broadcast) {
            this.$refs.modelSelect.broadcast(
              'ElSelectDropdown',
              'updatePopper',
            );
          }
        });
      }
    },
  },
  computed: {
    modelAvatar() {
      const o = this.options.find(o => o.modelId === this.currentValue);
      return this.currentValue.length && o.avatar.path
        ? avatarSrc(o.avatar.path)
        : defaultModelAvatar;
    },
  },
  methods: {
    handleChange(value) {
      this.$emit('change', value);
    },
    handleVisibleChange(value) {
      this.$emit('visible-change', value);
    },
    handleOptionClick(item) {
      const selectedOption = this.options.find(
        option => option.modelId === item.modelId,
      );
      if (selectedOption?.allowEdit === false && this.warning) {
        this.$message.warning(this.$t('modelAccess.publicWarning'));
      }
      this.$emit('option-click', item);
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/modelSelect.scss';
</style>
