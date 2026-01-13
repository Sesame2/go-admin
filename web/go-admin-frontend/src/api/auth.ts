import request from './request'
import type { LoginInput, LoginResponse } from '../types'

export const authAPI = {
    login(data: LoginInput): Promise<LoginResponse> {
        return request.post('/auth/login', data)
    },

    refresh(): Promise<LoginResponse> {
        return request.post('/auth/refresh')
    }
}
