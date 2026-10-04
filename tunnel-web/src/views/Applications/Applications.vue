<template>
    <div style="text-align: left;">
        <el-button plain @click="addHandlerClick">
            Add Applications
        </el-button>
    </div>
  
  <el-table :data="appList" stripe style="width: 100%">
    <el-table-column prop="id" label="id" width="120" />
    <el-table-column prop="name" label="name" width="180" />
    <el-table-column prop="type" label="type" width="180"/>
    <el-table-column prop="local_ip" label="local_ip" width="180" />
    <el-table-column prop="local_port" label="local_port" width="120" />
    <el-table-column prop="entry_domain" label="entry_domain" width="180" />
    <el-table-column prop="entry_port" label="entry_port" width="120" />
    <el-table-column prop="proxy" label="proxy" width="180" />
    <el-table-column prop="proxy_port" label="proxy_port" width="120" />
    <el-table-column fixed="right" label="Operations" min-width="120">
      <template #default="{row}">
        <el-button link type="primary" size="small" @click="editHandleClick(row)">
          Edit
        </el-button>
        <el-button
            link
            type="danger"
            size="small"
            @click="deleteHandleClick(row)"
        >
          Delete
        </el-button>
      </template>
    </el-table-column>
    
  </el-table>

  
  
  <el-dialog
    v-model="dialogFormVisible"
    :title="dialogTitle"
    width="500">
    <el-form :model="form" label-width="10rem">
            <input type="hidden" v-model="form.id" />
            <el-form-item label="Application Name">
                <el-input v-model="form.name" />
            </el-form-item>
            <el-form-item label="Application Type">
                <el-select v-model="form.type" placeholder="please select your type">
                    <el-option label="Http" value="http" />
                    <el-option label="SSH" value="ssh" />
                </el-select>
            </el-form-item>
            <el-form-item label="Local IP">
                <el-input v-model="form.local_ip" />
            </el-form-item>
            <el-form-item label="Local Port">
                <el-input v-model="form.local_port" />
            </el-form-item>
        </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="onCancel">Cancel</el-button>
        <el-button type="primary" @click="onSubmit">
          Confirm
        </el-button>
      </div>
    </template>
  </el-dialog>




</template>

<script lang="ts" setup>
import {reactive, ref, onBeforeMount} from 'vue';
import { getApplications,addApplication,app, editApplication, deleteApplication } from './Applications';
import {ElMessage, ElMessageBox} from "element-plus";


const dialogFormVisible = ref(false);

const form = reactive({
    id: 0,
    name: '',
    type: '',
    local_ip: '127.0.0.1',
    local_port: 0,
    
})

const appList = reactive([] as app[]);

const dialogTitle = ref('New Application');


onBeforeMount(() => {
    loadList();
});

const onSubmit = () => {
    if(form.id > 0) {
        editApplication(form as app).then(() => {
            loadList();
            dialogFormVisible.value = false;
            resetForm();
        }).catch((res) => {
            console.log(res);
        });
    } else {
        addApplication(form as app).then(() => {
            dialogFormVisible.value = false;
            loadList();
            resetForm();
        }).catch((res) => {
            console.log(res);
        });
    }
}


function loadList(){
   getApplications().then(function(res){
    appList.splice(0);
    for(let a of res.data) {
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
   }) 
}

const onCancel = () => {
    dialogFormVisible.value = false;
    resetForm();
}


const editHandleClick = (row: any) => {
    form.id = row.id;
    form.name = row.name;
    form.type = row.type;
    form.local_ip = row.local_ip;
    form.local_port = row.local_port;
    dialogFormVisible.value=true;
    dialogTitle.value = 'Edit Application';
}
const deleteHandleClick = (row: any) => {
  ElMessageBox.confirm(
      `确定要删除应用 "${row.name}" 吗？`,
      'Delete Application',
      {
        confirmButtonText: 'Confirm',
        cancelButtonText: 'Cancel',
        type: 'warning',
      }
  ).then(() => {
    deleteApplication(row.id).then(() => {
      ElMessage.success('删除成功');
      loadList();
    }).catch((res) => {
      console.log(res);
      ElMessage.error('删除失败');
    });
  }).catch(() => {
    // 用户取消删除
  });
};
const addHandlerClick = () => {
    dialogFormVisible.value = true;
    dialogTitle.value = 'New Application';
    resetForm();
}
function resetForm() {
    form.id = 0;
    form.name = '';
    form.type = '';
    form.local_ip = '127.0.0.1';
    form.local_port = 0;
}


</script>
