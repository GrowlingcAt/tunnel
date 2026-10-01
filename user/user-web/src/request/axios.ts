import axios,{ AxiosError } from "axios"
import { ElMessage } from 'element-plus'

const service  = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL
})
service.interceptors.request.use(
    function(config){
        //请求之前做点什么，比如添加令牌到请求头
        return config
    },
    function(error){
        //如果发送错误做点什么
        return Promise.reject(error);
    }
)
service.interceptors.response.use(
    function (response) {
        // 2xx 范围内的状态码都会触发该函数。
        // 对响应数据做点什么
        return response;
      }, function (error) {
        // 超出 2xx 范围的状态码都会触发该函数。
        // 对响应错误做点什么
        const axiosErr = error as AxiosError
        if (!axiosErr.response?.status) {
            ElMessage({
                showClose: true,
                message: axiosErr.message,
                type: 'error',
              })
        }else if(axiosErr.response?.status == 500){
            ElMessage({
                showClose: true,
                message: "服务器内部错误",
                type: 'error',
              })
        }else if(axiosErr.response?.status == 504){
            ElMessage({
                showClose: true,
                message: "网关超时",
                type: 'error',
              })
        }else if(axiosErr.response?.status == 413){
            ElMessage({
                showClose: true,
                message: "仅支持上传20m以内的图片",
                type: 'error',
              })
        }
        console.log(axiosErr.message)
        console.log(axiosErr.response?.status)
        return Promise.reject(error);
      }   
)
export default service