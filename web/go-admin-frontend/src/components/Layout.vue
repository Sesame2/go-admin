<template>
  <el-container class="layout-container">
    <el-header class="header">
      <div class="header-left">
        <span class="logo">📚 Go Admin</span>
      </div>
      <div class="header-right">
        <el-dropdown trigger="click" @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="32" class="avatar">{{ username.charAt(0).toUpperCase() }}</el-avatar>
            <span class="username">{{ username }}</span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout" :icon="SwitchButton">
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <el-container class="main-container">
      <el-aside width="220px" class="aside">
        <el-menu
          :default-active="activeMenu"
          router
          class="menu"
        >
          <el-menu-item index="/">
            <el-icon><FolderOpened /></el-icon>
            <span>知识库管理</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, SwitchButton, FolderOpened } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const username = ref(localStorage.getItem('username') || '用户')
const activeMenu = computed(() => route.path === '/' ? '/' : route.path)

const handleCommand = async (command: string) => {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })

      localStorage.removeItem('token')
      localStorage.removeItem('username')
      ElMessage.success('已退出登录')
      router.push('/login')
    } catch {
      // User cancelled
    }
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.logo {
  font-size: 20px;
  font-weight: 600;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: white;
}

.avatar {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.username {
  font-size: 14px;
}

.main-container {
  height: calc(100vh - 60px);
}

.aside {
  background: #fff;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.06);
}

.menu {
  height: 100%;
  border-right: none;
}

.menu :deep(.el-menu-item) {
  margin: 4px 8px;
  border-radius: 8px;
}

.menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.main {
  background: #f5f7fb;
  padding: 24px;
  overflow-y: auto;
}
</style>
