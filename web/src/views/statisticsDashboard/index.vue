<template>
  <div class="page-wrapper" style="padding-top: 20px">
    <div class="tabs" style="margin-top: 0">
      <div :class="['tab', { active: tabActive === 0 }]" @click="tabClick(0)">
        {{ $t('statisticsDashboard.model') }}
      </div>
    </div>

    <div v-if="tabActive === 0">
      <Model />
    </div>
  </div>
</template>

<script>
import Model from './components/model.vue';

export default {
  components: { Model },
  data() {
    return {
      radio: '',
      tabActive: 0,
    };
  },
  methods: {
    tabClick(status) {
      this.tabActive = status;
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tabs.scss';
</style>
