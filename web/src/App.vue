<template>
  <div class="app">
    <el-config-provider :locale="zhCn">
      <router-view v-slot="{ Component }">
        <component :is="Component" />
      </router-view>
    </el-config-provider>
  </div>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import locale from "element-plus/es/locale/lang/zh-cn";
const zhCn = ref(locale);
</script>

<style lang="scss" scoped>
.app {
  font-family: "Avenir", Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  height: 100%;
  width: 100%;
  background-color: #f5f5f9;
}
</style>
