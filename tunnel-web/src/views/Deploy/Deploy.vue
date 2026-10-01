<template>
  <el-tabs type="border-card">
    <el-tab-pane label="Tunnel Client cofig">
      <pre>config.yaml</pre>
            <div style="position: relative">
                <pre>{{ yamlconfig }}<el-icon class="copyicon" size="1.5rem"  @click="handleCopy"><DocumentCopy /></el-icon></pre>
            </div> 
    </el-tab-pane>
    <el-tab-pane label="Tunnel Client Install">
        <h3 style="text-align: left">docker</h3>
            <el-divider border-style="dashed" />
            <pre># 保存配置到 <br/>/home/tunnel-client/config.yaml</pre>
            <pre># 拉取镜像<br/>docker pull quay.io/0voice/tunnel-client:0.62.1</pre>
            <pre># 启动客户端<br/>docker run -d --name tunnel-client --network host  -v /home/tunnel-client/config.yaml:/app/config.yaml quay.io/0voice/tunnel-client:0.62.1</pre>
    </el-tab-pane>
    <el-tab-pane label="Visit Application">
      <el-table :data="appList" style="width: 100%">
                <el-table-column prop="name" label="Name" width="180" />
                <el-table-column prop="type" label="Type" width="180" />
                <el-table-column prop="entry_domain" label="Entry" width="280" />
                <el-table-column prop="entry_port" label="Entry Port" width="180" />
                <el-table-column prop="proxy" label="Proxy" width="280" />
                <el-table-column prop="proxy_port" label="Proxy Port" />
            </el-table> 
    </el-tab-pane>
  </el-tabs>
</template>

<script lang="ts" setup>
import { onBeforeMount,reactive,ref } from 'vue';
import { deploy } from './Deploy.ts';
import * as YAML from 'js-yaml';
import { ElMessage } from 'element-plus';
import {DocumentCopy} from '@element-plus/icons-vue';
let yamlconfig = ref('');
interface app {
    id: number;
    name: string;
    type: string;
    local_ip: string;
    local_port: number;
    entry_domain: string;
    entry_port: number;
    proxy: string;
    proxy_port: number;
}
let appList = reactive([] as app[]);
onBeforeMount(() => {
    deploy().then((res) => {
        const parseObject = YAML.load(res.data.client_config);
        yamlconfig.value = YAML.dump(parseObject);
        appList.splice(0);
        for(let a of res.data.apps) {
        let item = {
            id: a.id,
            name: a.name,
            type: a.type,
            local_ip: a.local_ip,
            local_port: a.local_port,
            entry_domain: a.entry_domain,
            entry_port: a.entry_port,
            proxy: a.proxy,
            proxy_port: a.proxy_port,
        }
        appList.push(item);
        }

    }).catch((res) => {
        console.log(res);
    });
})

function handleCopy() {
    navigator.clipboard
        .writeText(yamlconfig.value)
        .then(function () {
            ElMessage({
                message: "已复制到剪切板",
                type: "success",
            });
        })
        .catch(function () {
            ElMessage.error("复制失败，请手动复制文本框内链接");
        });
}
</script>
<style scoped>
pre {
    background-color: #f4f4f4;
    padding: 10px;
    border-radius: 5px;
    font-family: Consolas, monospace;
    line-height: 1.5;
    white-space: pre-wrap;
    text-align: left;
    font-size: 1.2rem;
}

.copyicon {
    position: absolute;
    top: 1rem;
    right: 1rem;
    cursor: pointer;
}
</style>