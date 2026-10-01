<template>
  <div class="common-layout">
    <el-container>
      <el-aside width="200px"><siderbar-menu></siderbar-menu></el-aside>
      <el-main>
        <RouterView />
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router';
import SiderbarMenu from '../../components/SiderbarMenu.vue';
import { onBeforeMount } from 'vue';
import { useRouter } from 'vue-router';
let router = useRouter();
onBeforeMount(()=>{
  router.push("/apps");
  router.push("/deploy")
});

</script>
