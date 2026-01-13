<template>
  <div class="document-page">
    <el-page-header @back="goBack">
      <template #content>
        <span class="page-title">📄 {{ kbName }} - 文档管理</span>
      </template>
      <template #extra>
        <el-button type="primary" :icon="Upload" @click="showUploadDialog = true">上传文档</el-button>
        <el-button :type="autoRefresh ? 'success' : 'default'" @click="toggleAutoRefresh">
          {{ autoRefresh ? '停止刷新' : '自动刷新' }}
        </el-button>
      </template>
    </el-page-header>

    <el-card class="document-card" shadow="hover">
      <el-table :data="documents" v-loading="listLoading" stripe>
        <el-table-column prop="filename" label="文档名称" min-width="200">
          <template #default="{ row }">
            <el-tooltip :content="row.filename" placement="top">
              <span class="doc-name">📄 {{ row.filename }}</span>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column prop="file_size" label="大小" width="100">
          <template #default="{ row }">{{ formatSize(row.file_size) }}</template>
        </el-table-column>

        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="chunk_count" label="分块数" width="80" align="center" />

        <el-table-column prop="created_at" label="上传时间" width="180">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>

        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'pending' || row.status === 'failed'"
              type="primary"
              size="small"
              @click="handleParse(row)"
            >解析</el-button>
            <el-button type="info" size="small" @click="handleDownload(row)">下载</el-button>
            <el-button type="success" size="small" @click="handleViewDetail(row)">详情</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="total > pageSize"
        v-model:current-page="page"
        :page-size="pageSize"
        layout="total, prev, pager, next"
        :total="total"
        class="pagination"
        @current-change="loadDocuments"
      />
    </el-card>

    <!-- Upload Dialog -->
    <el-dialog v-model="showUploadDialog" title="上传文档" width="500px">
      <el-upload
        ref="uploadRef"
        drag
        :auto-upload="false"
        :on-change="handleFileChange"
        :limit="1"
        accept=".pdf,.txt,.md,.docx,.doc"
      >
        <el-icon class="upload-icon"><UploadFilled /></el-icon>
        <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
        <template #tip>
          <div class="el-upload__tip">支持 PDF, TXT, MD, DOCX 等格式</div>
        </template>
      </el-upload>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" :loading="uploadLoading" :disabled="!selectedFile" @click="handleUpload">
          上传
        </el-button>
      </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="文档详情" width="600px">
      <el-descriptions v-if="currentDocument" :column="1" border>
        <el-descriptions-item label="文档ID">
          <el-tag size="small">{{ currentDocument.id }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="名称">{{ currentDocument.filename }}</el-descriptions-item>
        <el-descriptions-item label="大小">{{ formatSize(currentDocument.file_size) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentDocument.status)">
            {{ getStatusText(currentDocument.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="分块数">{{ currentDocument.chunk_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="上传时间">{{ formatDate(currentDocument.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDate(currentDocument.updated_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="currentDocument.error_message" label="错误信息">
          <el-text type="danger">{{ currentDocument.error_message }}</el-text>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type UploadFile } from 'element-plus'
import { Upload, UploadFilled } from '@element-plus/icons-vue'
import { documentAPI, knowledgeBaseAPI } from '../api'
import type { Document, DocumentStatus } from '../types'

const route = useRoute()
const router = useRouter()

const kbId = ref(route.params.kbId as string)
const kbName = ref('')
const listLoading = ref(false)
const uploadLoading = ref(false)
const showUploadDialog = ref(false)
const showDetailDialog = ref(false)
const autoRefresh = ref(false)
let autoRefreshTimer: number | null = null

const documents = ref<Document[]>([])
const currentDocument = ref<Document | null>(null)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const selectedFile = ref<File | null>(null)

const statusMap: Record<DocumentStatus, { text: string; type: 'info' | 'warning' | 'primary' | 'success' | 'danger' }> = {
  'uploaded': { text: '已上传', type: 'info' },
  'pending': { text: '待处理', type: 'warning' },
  'downloading': { text: '下载中', type: 'primary' },
  'parsing': { text: '解析中', type: 'primary' },
  'chunking': { text: '分块中', type: 'primary' },
  'embedding': { text: '向量化中', type: 'primary' },
  'completed': { text: '已完成', type: 'success' },
  'failed': { text: '失败', type: 'danger' }
}

const loadKnowledgeBase = async () => {
  try {
    const res = await knowledgeBaseAPI.getDetail(kbId.value)
    kbName.value = res.name
  } catch (error) {
    console.error('Failed to load KB:', error)
  }
}

const loadDocuments = async () => {
  listLoading.value = true
  try {
    const res = await documentAPI.getList(kbId.value, { page: page.value, page_size: pageSize.value })
    documents.value = res.documents || []
    total.value = res.total || 0
  } catch (error) {
    console.error('Failed to load documents:', error)
  } finally {
    listLoading.value = false
  }
}

const handleFileChange = (file: UploadFile) => {
  selectedFile.value = file.raw || null
}

const handleUpload = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请选择文件')
    return
  }
  uploadLoading.value = true
  try {
    await documentAPI.upload(kbId.value, selectedFile.value)
    ElMessage.success('上传成功')
    showUploadDialog.value = false
    selectedFile.value = null
    loadDocuments()
  } catch (error) {
    console.error('Upload failed:', error)
  } finally {
    uploadLoading.value = false
  }
}

const handleParse = async (doc: Document) => {
  try {
    await ElMessageBox.confirm(`确定要解析文档"${doc.name}"吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info'
    })
    await documentAPI.parse(doc.id)
    ElMessage.success('已提交解析任务')
    if (!autoRefresh.value) toggleAutoRefresh()
    loadDocuments()
  } catch (error) {
    if (error !== 'cancel') console.error('Parse failed:', error)
  }
}

const handleDownload = async (doc: Document) => {
  try {
    const res = await documentAPI.getDownloadUrl(doc.id)
    if (res.url) {
      window.open(res.url, '_blank')
      ElMessage.success('开始下载')
    } else {
      ElMessage.error('获取下载URL失败')
    }
  } catch (error) {
    console.error('Download failed:', error)
  }
}

const handleViewDetail = async (doc: Document) => {
  try {
    const res = await documentAPI.getDetail(doc.id)
    currentDocument.value = res
    showDetailDialog.value = true
  } catch (error) {
    console.error('Get detail failed:', error)
  }
}

const handleDelete = async (doc: Document) => {
  try {
    await ElMessageBox.confirm(`确定要删除文档"${doc.name}"吗？此操作不可恢复。`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await documentAPI.delete(doc.id)
    ElMessage.success('删除成功')
    loadDocuments()
  } catch (error) {
    if (error !== 'cancel') console.error('Delete failed:', error)
  }
}

const toggleAutoRefresh = () => {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    autoRefreshTimer = window.setInterval(() => loadDocuments(), 5000)
    ElMessage.info('已开启自动刷新')
  } else {
    if (autoRefreshTimer) {
      clearInterval(autoRefreshTimer)
      autoRefreshTimer = null
    }
    ElMessage.info('已停止自动刷新')
  }
}

const goBack = () => router.push({ name: 'KnowledgeBase' })

const getStatusText = (status: DocumentStatus) => statusMap[status]?.text || status
const getStatusType = (status: DocumentStatus) => statusMap[status]?.type || 'info'

const formatSize = (bytes: number) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadKnowledgeBase()
  loadDocuments()
})

onUnmounted(() => {
  if (autoRefreshTimer) clearInterval(autoRefreshTimer)
})
</script>

<style scoped>
.document-page {
  min-height: 100%;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
}

.document-card {
  margin-top: 24px;
}

.doc-name {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pagination {
  margin-top: 20px;
  justify-content: center;
}

.upload-icon {
  font-size: 67px;
  color: #667eea;
  margin-bottom: 16px;
}
</style>
