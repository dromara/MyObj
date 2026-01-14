<!--
  路由加载组件（已合并到 Loading 组件，此文件仅为向后兼容）
  @deprecated 请使用 <Loading type="route" />
-->
<script setup lang="ts">
  import Loading from '../Loading/index.vue'

  interface Props {
    loading?: boolean
    text?: string
  }

  defineProps<Props>()
</script>

<template>
  <Loading :loading="loading" :text="text" type="route" />
</template>
