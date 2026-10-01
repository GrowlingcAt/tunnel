<template>
  <el-row class="tac">
    
    <el-col :span="24">
      <h5 class="mb-2">Navigator</h5>
      <el-menu
        active-text-color="#ffd04b"
        background-color="#545c64"
        class="el-menu-vertical-demo"
        default-active="2"
        text-color="#fff"
        @open="handleOpen"
        @close="handleClose"
        router
      >
        <el-menu-item index="/apps">
          <el-icon><icon-menu /></el-icon>
          <span>Applications</span>
        </el-menu-item>
        <el-menu-item index="/deploy">
          <el-icon><setting /></el-icon>
          <span>Deploy</span>
        </el-menu-item>
      </el-menu>
    </el-col>
  </el-row>
</template>

<script lang="ts" setup>

const handleOpen = (key: string, keyPath: string[]) => {
  console.log(key, keyPath)
}
const handleClose = (key: string, keyPath: string[]) => {
  console.log(key, keyPath)
}
</script>
