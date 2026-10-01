<template>
<div>
  <el-row>
    <el-col>
      <el-card :body-style="{ padding: '0px' }">
        <img style="margin-bottom: -15px;" :src="data.loginMethods.wx_qrcode.qr_code_url.value" class="image" />
        <div >
          <span style="color:#999;font-size: 1rem;">微信扫码</span>
          <div class="bottom">
            <a :href="data.loginMethods.github.value">
              <svg width="25" height="24" aria-hidden="true" viewBox="0 0 16 16" version="1.1" data-view-component="true" class="octicon octicon-mark-github v-align-middle color-fg-default">
                <path d="M8 0c4.42 0 8 3.58 8 8a8.013 8.013 0 0 1-5.45 7.59c-.4.08-.55-.17-.55-.38 0-.27.01-1.13.01-2.2 0-.75-.25-1.23-.54-1.48 1.78-.2 3.65-.88 3.65-3.95 0-.88-.31-1.59-.82-2.15.08-.2.36-1.02-.08-2.12 0 0-.67-.22-2.2.82-.64-.18-1.32-.27-2-.27-.68 0-1.36.09-2 .27-1.53-1.03-2.2-.82-2.2-.82-.44 1.1-.16 1.92-.08 2.12-.51.56-.82 1.28-.82 2.15 0 3.06 1.86 3.75 3.64 3.95-.23.2-.44.55-.51 1.07-.46.21-1.61.55-2.33-.66-.15-.24-.6-.83-1.23-.82-.67.01-.27.38.01.53.34.19.73.9.82 1.13.16.45.68 1.31 2.69.94 0 .67.01 1.3.01 1.49 0 .21-.15.45-.55.38A7.995 7.995 0 0 1 0 8c0-4.42 3.58-8 8-8Z"></path>
              </svg>
            </a>
            <a style="margin-left: 0.5rem;" :href="data.loginMethods.gitlab.value" >
              <svg class="tanuki-logo" width="25" height="24" viewBox="0 0 25 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path class="tanuki-shape tanuki" d="m24.507 9.5-.034-.09L21.082.562a.896.896 0 0 0-1.694.091l-2.29 7.01H7.825L5.535.653a.898.898 0 0 0-1.694-.09L.451 9.411.416 9.5a6.297 6.297 0 0 0 2.09 7.278l.012.01.03.022 5.16 3.867 2.56 1.935 1.554 1.176a1.051 1.051 0 0 0 1.268 0l1.555-1.176 2.56-1.935 5.197-3.89.014-.01A6.297 6.297 0 0 0 24.507 9.5Z" fill="#E24329"></path>
                <path class="tanuki-shape right-cheek" d="m24.507 9.5-.034-.09a11.44 11.44 0 0 0-4.56 2.051l-7.447 5.632 4.742 3.584 5.197-3.89.014-.01A6.297 6.297 0 0 0 24.507 9.5Z" fill="#FC6D26"></path>
                <path class="tanuki-shape chin" d="m7.707 20.677 2.56 1.935 1.555 1.176a1.051 1.051 0 0 0 1.268 0l1.555-1.176 2.56-1.935-4.743-3.584-4.755 3.584Z" fill="#FCA326"></path>
                <path class="tanuki-shape left-cheek" d="M5.01 11.461a11.43 11.43 0 0 0-4.56-2.05L.416 9.5a6.297 6.297 0 0 0 2.09 7.278l.012.01.03.022 5.16 3.867 4.745-3.584-7.444-5.632Z" fill="#FC6D26"></path>
                </svg>
            </a>
          </div>
        </div>
      </el-card>
    </el-col>
  </el-row>
</div>
</template>
<script lang="ts" setup>
import { onBeforeMount,ref } from 'vue';
import { ElMessage } from 'element-plus'
import {getUrlParameter} from '../../utils/utils'
import {loginMethods,loginOfficialCallback} from './login'
let data = {
  loginMethods : {
    github: ref(""),
    gitlab: ref(""),
    wx_qrcode:{
      expire_seconds: ref(0),
      ticket: ref(""),
      qr_code_url: ref(""),
    }
  }
}
onBeforeMount(() =>{
  const sys = getUrlParameter("sys")
  loginMethods({"sys":sys}).then(function(res){
    data.loginMethods.github.value = res.data.github
    data.loginMethods.gitlab.value = res.data.gitlab
    data.loginMethods.wx_qrcode.expire_seconds.value = res.data.wx_qrcode.expire_seconds
    data.loginMethods.wx_qrcode.ticket.value = res.data.wx_qrcode.ticket
    data.loginMethods.wx_qrcode.qr_code_url.value = res.data.wx_qrcode.qr_code_url

    let count = 0;
    const timer = setInterval(() => {
      loginOfficialCallback({sys:sys,ticket:data.loginMethods.wx_qrcode.ticket.value}).then((res) => {
        if (res.data && res.data.redirect_url != "") {
          clearInterval(timer)
          window.location.href=res.data.redirect_url
        }
        count ++;
        if (count >= 180) {
          clearInterval(timer)
          ElMessage({
            showClose: true,
            message: "二维码已过期，请刷新页面",
            type: 'error',
          })
        }
      }).catch((res)=>{
        clearInterval(timer)
        ElMessage({
            showClose: true,
            message: res.message,
            type: 'error',
        }) 
      })
    },1000)

  }).catch(function(res){
    ElMessage({
            showClose: true,
            message: res.message,
            type: 'error',
        }) 
  })
})

</script>
<style>
.time {
  font-size: 12px;
  color: #999;
}

.bottom {
  background-color:rgb(239, 241, 242) ;
  padding:0.5rem;
  line-height: 0.7rem;
  display:flex;
  justify-content: flex-start;
  align-items: center;
}

.button {
  padding: 0;
  min-height: auto;
}

.image {
  width: 16rem;
  height: 16rem;
  display: block;
}
</style>