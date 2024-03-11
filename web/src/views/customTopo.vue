<!-- 自定义topo图页面 -->
<template>
  <div style="width: 96%;margin:0 auto;height: 100%; display: flex;align-items: center;">
    <div class="form">
      <el-scrollbar height="100%">
        <el-form ref="formRef" :rules="rules" :model="customForm" label-width="auto">
          <el-form-item prop="conf_name" label="配置名称：">
            <el-input v-model="customForm.conf_name" />
          </el-form-item>
          <el-form-item prop="conf_version" label="配置版本：">
            <el-input v-model="customForm.conf_version" />
          </el-form-item>
          <el-form-item label="配置描述：">
            <el-input v-model="customForm.description" type="textarea" />
          </el-form-item>
          <el-form-item prop="batchId" label="机器批次：">
            <el-select v-model="customForm.batchName" placeholder="请选择批次" clearable @change="handlebatchDetail">
              <el-option v-for="item in batchs" :label="item.name" :value="item.id + '~' + item.name" />
            </el-select>
          </el-form-item>
          <el-form-item v-for="(ruleItem, index) in customForm.node_rules" :label="'节点规则' + (index + 1) + '：'">
            <el-space direction="vertical" alignment="flex-start">
              <el-select v-model="ruleItem[0].rule_condition.ip" placeholder="请选择机器" clearable @change="handleSelectHost($event,index)">
                <el-option v-for="item in hosts" :label="item.ip" :value="item.ip + '~'+item.uuid" />
              </el-select>
              <el-select v-model="ruleItem[1].rule_type" placeholder="请选择规则类型" @change="handleRuleType($event,index)">
                <el-option v-for="item in rule_types" :label="item.type" :value="item.type" />
              </el-select>


              <el-input v-if="ruleItem[1].rule_type === 'process'" v-model="ruleItem[1].rule_condition.name"
                placeholder="请输入节点具体内容" />
              <el-input v-if="ruleItem[1].rule_type === 'tag'" v-model="ruleItem[1].rule_condition.tag_name"
                placeholder="请输入节点具体内容" />
              <el-checkbox-group v-if="ruleItem[1].rule_type === 'resource'"
                v-model="ruleItem[1].rule_condition.resource">
                <el-checkbox label="cpu" value="cpu" disabled checked/>
                <el-checkbox label="disk" value="disk" disabled checked/>
                <el-checkbox label="iface" value="iface" disabled checked/>
                
              </el-checkbox-group>
              <el-button @click="delNodeRule(index)" link type="danger">- 删除规则</el-button>

            </el-space>
          </el-form-item>
          <el-form-item>
            <el-button @click="addNodeRule" link type="primary">+ 新增规则</el-button>
          </el-form-item>
          <el-form-item>
            <div style="width:90%;display: flex; justify-content: flex-end;">
              <el-button @click="resetForm(formRef)">重置</el-button>
              <el-button v-if="showTopo" @click="showTopo = !showTopo">查看json</el-button>
              <el-button v-if="isEdit" @click="submitForm(formRef,'edit')">更新</el-button>
              <el-button type="primary" @click="submitForm(formRef,'create')">创建</el-button>
            </div>
          </el-form-item>
        </el-form>
      </el-scrollbar>
    </div>
    <div class="topo">
      <PGTopo v-if="showTopo" style="width: 100%;height: 100%;" :graph_mode="graphMode" :time_interval="timeInterval" />
      <el-scrollbar height="100%" v-else> 
        <vue-json-pretty  :data="customForm" showLength/>
      </el-scrollbar>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { getBatchList, getBatchDetail,addConfList,updateConfList, getCustomTopo } from '@/request/api';
import PGTopo from '@/components/PGTopo.vue';
import { useTopoStore } from '@/stores/topo';
import type { TopoCustomFormType } from '@/types';
import VueJsonPretty from 'vue-json-pretty';
import 'vue-json-pretty/lib/styles.css';
import { useConfigStore } from '@/stores/config';
const formRef = ref<FormInstance>()
const graphMode = ref('default');
const timeInterval = ref('10');
const isEdit = ref(false); // 是否可编辑配置
const configId = ref(0); // 编辑文件接口参数
const showTopo = ref(false);

let customForm = reactive<TopoCustomFormType>({
  conf_name: '',
  conf_version:'',
  conf_time: Date.now() + '',
  batchId: 0,
  batchName:'',
  node_rules: [[{ rule_condition: {}, rule_type: 'host' }, { rule_condition: {}, rule_type: '' }]],
  description: '',
})
interface CustomForm {
  conf_name: string
  batchId: number
}

const rules = reactive<FormRules<CustomForm>>({
  conf_name: [
    { required: true, message: '请输入配置名称', trigger: 'blur' },
  ],
  batchId: [
    {
      required: true,
      message: '请选择批次',
      trigger: 'blur',
    },
  ],
})

let batchs = reactive([
  {
    id: 0,
    name: ''
  }
]);
let hosts = reactive([
  {
    uuid: '',
    ip: ''
  }
])

const rule_types = [
  {
    type: 'process',
    keyname: 'name',
  },
  {
    type: 'tag',
    keyname: 'tag_name',
  },
  {
    type: 'resource',
    keyname: '',
  },
]

onMounted(() => {
  getBatchList().then(res => {
    if (res.data.code == 200) {
      batchs = res.data.data;
    }
  })
})

// 选择机器
const handleSelectHost = (e:any,index:number) => {
  let hostInfo = e.split('~');
  customForm.node_rules[index][0].rule_condition.ip=hostInfo[0]
  customForm.node_rules[index][0].rule_condition.uuid=hostInfo[1]
}
// 处理机器类型
const handleRuleType = (type:string,index: number) => {
  customForm.node_rules[index][1].rule_condition = {}
  if (type === 'resource') {
    customForm.node_rules[index][1].rule_condition = {"resource":["cpu","disk","iface"]}
  }
}

// 获取批次所有主机
const handlebatchDetail = (e: string) => {
  let baychInfo = e.split('~');
  customForm.batchId = Number(baychInfo[0]);
  customForm.batchName = baychInfo[1];
  getBatchDetail({ batchId: customForm.batchId }).then(res => {
    if (res.data.code === 200) {
      hosts = res.data.data;
    }
  })
}
// 删除规则
const delNodeRule = (index: number) => {
     customForm.node_rules.splice(index, 1)
}
// 添加规则
const addNodeRule = () => {
  customForm.node_rules.push([
    { rule_condition: {}, rule_type: 'host' }, { rule_condition: {}, rule_type: '' }
  ])

}

// 获取topo数据
const getTopo = () => {
  getCustomTopo({ id: configId.value }).then(res => {
    if (res.data.code === 200) {
      useTopoStore().topo_type = 'comb';
      useTopoStore().topo_data = res.data.data;
    }
  })
}

// 创建/编辑 配置文件
const submitForm = (formEl: FormInstance | undefined,type:string) => {
  if (!formEl) return
  formEl.validate((valid) => {
    if (valid) {
      showTopo.value = true;
      if (type === 'create') {
        // 新增
        addConfList(customForm).then(res => {
          if (res.data.code === 200) {
            configId.value = res.data.data;
            ElMessage.success(res.data.msg);
            isEdit.value = true;
            getTopo();
          } else {
            ElMessage.error(res.data.msg);
          }
        })
      } else {
        // 编辑
        updateConfList({id:configId.value,topo_configuration:customForm}).then(res => {
          if (res.data.code === 200) {
            configId.value = res.data.data;
            getTopo();
          } else {
            ElMessage.error(res.data.msg);
          }
        })
      }
    } else {
      console.log('error submit!')
      return false
    }
  })
}

const resetForm = (formEl: FormInstance | undefined) => {
  if (!formEl) return
  formEl.resetFields();
  customForm.node_rules = [[{ rule_condition: {}, rule_type: 'host' }, { rule_condition: {}, rule_type: '' }]]
}
let formInfo = ref({});
watch(() => customForm, (newForm) => {
  if (newForm) {
    formInfo.value = JSON.parse(JSON.stringify(newForm));
  }
}, {
  immediate: true,
  deep: true
})

watch(() => useConfigStore().topo_config, (newConfig) => {
  if (newConfig) {
    isEdit.value = true;
    let currentConf = JSON.parse(JSON.stringify(newConfig));
    customForm.conf_name = currentConf.conf_name;
    customForm.conf_version = currentConf.conf_version;
    customForm.batchId = currentConf.batchId;
    customForm.node_rules = currentConf.node_rules;
    customForm.description = currentConf.description;
    configId.value = currentConf.id;
    setTimeout(() => {
      customForm.batchName = batchs.filter(item => item.id === customForm.batchId)[0].name;
    },100)
  }
},{immediate:true,deep:true})
</script>

<style scoped lang="scss">
.form {
  width: 30%;
  height: 100%;
  padding-right: 2px;
  border-right: 1px solid #c8c9cc;
}

.topo {
  width:68%;
  height: 100%;
  overflow: hidden;
}
</style>