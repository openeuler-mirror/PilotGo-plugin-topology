<!--
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhao_zhen_fang <zhaozhenfang@kylinos.cn> 
 * Date: Mon Oct 9 11:19:00 2023 +0800
-->
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
