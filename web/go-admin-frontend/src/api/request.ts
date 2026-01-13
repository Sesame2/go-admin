import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

const request: AxiosInstance = axios.create({
    baseURL: 'http://localhost:8080/api',
    timeout: 60000
})

// Request interceptor - add auth token
request.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem('token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

// Response interceptor - handle errors and return data directly
request.interceptors.response.use(
    (response: AxiosResponse) => {
        return response.data
    },
    (error) => {
        if (error.response) {
            const { status, data } = error.response

            if (status === 401) {
                ElMessage.error('登录已过期，请重新登录')
                localStorage.removeItem('token')
                localStorage.removeItem('username')
                router.push('/login')
            } else {
                ElMessage.error(data?.message || data?.error || '请求失败')
            }
        } else {
            ElMessage.error('网络错误，请检查网络连接')
        }

        return Promise.reject(error)
    }
)

export default request
