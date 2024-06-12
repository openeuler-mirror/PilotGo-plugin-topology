<!-- elk插件展示日志页面 -->
<template>
  <div class="log_containt">

  </div>

</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { getLogData } from '../../request/elk'

onMounted(() => {
  getLogs();
})

const getLogs = () => {
  getLogData({}).then((res: any) => {

  })
}

</script>


<style scoped lang="scss">
.log_containt {
  width: 96%;
  height: 400px;
  position: fixed;
  bottom: 0;
  border: 1px solid red;
  background-color: #fff;

}
</style>