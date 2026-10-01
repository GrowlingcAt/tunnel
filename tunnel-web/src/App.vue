<template>
  <div class="common-layout">
    <el-container>
      <el-header class = header>
        <el-avatar :size="60" src="user.avatar">
      <img src="https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png"/>
    </el-avatar>
      </el-header>
      <el-main style ="margin: 0; padding: 0;">
        <RouterView />
      </el-main>
    </el-container>
  </div>
</template>

<script>
import { reactive,onBeforeMount } from 'vue';
import {getCookieValue} from './utils/utils.ts';
import { useRouter} from 'vue-router';
const router = useRouter();
let user = reactive({
  name: '',
  avatar: '',
  id: 0,
});

onBeforeMount(()=>{
  let access_token = getCookieValue('sso_0voice_access_token');
  if(!access_token){
    window.location.href = import.meta.env.VITE_USER_CENTER;
  }else{
    let userInfoStr = atob(access_token.split('.')[1]);
    user = JSON.parse(userInfoStr);
    router.push("/")
  }
})
</script>


<style scoped>
.header{
  display: flex;
  align-items: center;
  justify-content: flex-end;
  background-color: grey;
}

.logo {
  height: 6em;
  padding: 1.5em;
  will-change: filter;
  transition: filter 300ms;
}
.logo:hover {
  filter: drop-shadow(0 0 2em #646cffaa);
}
.logo.vue:hover {
  filter: drop-shadow(0 0 2em #42b883aa);
}
</style>
