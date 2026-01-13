<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="logo">📚</div>
        <h1>Go Admin</h1>
        <p>知识库管理系统</p>
      </div>

      <el-form 
        ref="formRef" 
        :model="form" 
        :rules="rules" 
        class="login-form"
        @submit.prevent="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            :prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="login-btn"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '登录' }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { authAPI } from '../api'
import type { LoginInput } from '../types'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive<LoginInput>({
  username: '',
  password: ''
})

const rules: FormRules<LoginInput> = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const res = await authAPI.login(form)
        if (res.token) {
          localStorage.setItem('token', res.token)
          localStorage.setItem('username', form.username)
          ElMessage.success('登录成功')
          router.push('/')
        } else {
          ElMessage.error('登录异常：未获取到令牌')
        }
      } catch (error) {
        console.error('Login failed:', error)
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 420px;
  background: white;
  border-radius: 16px;
  padding: 48px 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.logo {
  font-size: 48px;
  margin-bottom: 16px;
}

.login-header h1 {
  font-size: 28px;
  font-weight: 600;
  color: #1a1a2e;
  margin: 0 0 8px;
}

.login-header p {
  color: #6b7280;
  font-size: 14px;
  margin: 0;
}

.login-form {
  margin-top: 32px;
}

.login-form :deep(.el-input__wrapper) {
  padding: 8px 15px;
  border-radius: 8px;
}

.login-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 8px;
  margin-top: 8px;
}
</style>
