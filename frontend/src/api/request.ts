import axios, { type AxiosInstance, type InternalAxiosRequestConfig } from 'axios'
import { getToken, removeToken } from '../utils/auth'
import { message } from 'ant-design-vue'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'

const service: AxiosInstance = axios.create({
    baseURL,
    timeout: 10000
})

// 请求拦截
service.interceptors.request.use(
    (config: InternalAxiosRequestConfig ) => {
        const token = getToken()
        if (token && config.headers) {
            config.headers['Authorization'] = `Bearer ${token}`
        }
        return config
    },
    (error) => Promise.reject(error)
)

// 响应拦截
service.interceptors.response.use(
    (response) => {
        const data = response.data
        if (data && data.code && data.code !== 200) {
            message.error(data.msg || '请求出错')
            return Promise.reject(new Error(data.msg || 'Error'))
        }
        return data
    },
    (error) => {
        const msg = error.response?.data?.msg || '请求失败'
        message.error(msg)
        return Promise.reject(error)
    }
)

export default service
