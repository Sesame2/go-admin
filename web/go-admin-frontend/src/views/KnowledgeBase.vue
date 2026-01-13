<template>
  <div class="knowledge-base-page">
    <div class="page-header">
      <h2>📚 知识库管理</h2>
      <el-button type="primary" :icon="Plus" @click="showCreateDialog = true">
        新建知识库
      </el-button>
    </div>

    <el-row :gutter="24">
      <!-- Knowledge Base List -->
      <el-col :span="10">
        <el-card class="list-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>知识库列表</span>
              <el-input
                v-model="searchKeyword"
                placeholder="搜索知识库"
                :prefix-icon="Search"
                clearable
                style="width: 200px"
              />
            </div>
          </template>

          <div v-loading="listLoading" class="kb-list">
            <el-empty v-if="filteredKnowledgeBases.length === 0" description="暂无知识库" />
            
            <div
              v-for="kb in filteredKnowledgeBases"
              :key="kb.id"
              :class="['kb-item', { active: selectedKb?.id === kb.id }]"
              @click="selectKnowledgeBase(kb)"
            >
              <div class="kb-icon">📖</div>
              <div class="kb-info">
                <div class="kb-title">{{ kb.name }}</div>
                <div class="kb-desc">{{ kb.description || '暂无描述' }}</div>
                <div class="kb-meta">创建于 {{ formatDate(kb.created_at) }}</div>
              </div>
            </div>
          </div>

          <el-pagination
            v-if="total > pageSize"
            v-model:current-page="page"
            :page-size="pageSize"
            layout="prev, pager, next"
            :total="total"
            small
            class="pagination"
            @current-change="loadKnowledgeBases"
          />
        </el-card>
      </el-col>

      <!-- Knowledge Base Detail -->
      <el-col :span="14">
        <el-card v-if="selectedKb" class="detail-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>{{ selectedKb.name }}</span>
              <div class="actions">
                <el-button size="small" :icon="Edit" @click="openEditDialog">编辑</el-button>
                <el-button size="small" type="danger" :icon="Delete" @click="handleDelete">删除</el-button>
              </div>
            </div>
          </template>

          <el-descriptions :column="1" border>
            <el-descriptions-item label="知识库ID">
              <el-tag size="small">{{ selectedKb.id }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="名称">{{ selectedKb.name }}</el-descriptions-item>
            <el-descriptions-item label="描述">{{ selectedKb.description || '暂无描述' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatDate(selectedKb.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="更新时间">{{ formatDate(selectedKb.updated_at) }}</el-descriptions-item>
          </el-descriptions>

          <div class="action-buttons">
            <el-button type="primary" :icon="Document" @click="goToDocuments">
              📄 管理文档
            </el-button>
            <el-button type="success" :icon="ChatLineRound" @click="goToRetrieval">
              💬 开始问答
            </el-button>
          </div>
        </el-card>

        <el-empty v-else description="请选择一个知识库" class="empty-detail" />
      </el-col>
    </el-row>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="新建知识库" width="480px">
      <el-form ref="createFormRef" :model="createForm" :rules="formRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入知识库名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" type="textarea" :rows="3" placeholder="请输入知识库描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- Edit Dialog -->
    <el-dialog v-model="showEditDialog" title="编辑知识库" width="480px">
      <el-form ref="editFormRef" :model="editForm" :rules="formRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="handleUpdate">更新</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Search, Edit, Delete, Document, ChatLineRound } from '@element-plus/icons-vue'
import { knowledgeBaseAPI } from '../api'
import type { KnowledgeBase, CreateKnowledgeBaseInput } from '../types'

const router = useRouter()

const listLoading = ref(false)
const createLoading = ref(false)
const editLoading = ref(false)
const showCreateDialog = ref(false)
const showEditDialog = ref(false)

const knowledgeBases = ref<KnowledgeBase[]>([])
const selectedKb = ref<KnowledgeBase | null>(null)
const searchKeyword = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const createFormRef = ref<FormInstance>()
const editFormRef = ref<FormInstance>()

const createForm = reactive<CreateKnowledgeBaseInput>({
  name: '',
  description: ''
})

const editForm = reactive<CreateKnowledgeBaseInput>({
  name: '',
  description: ''
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入知识库名称', trigger: 'blur' },
    { min: 2, max: 50, message: '名称长度在 2 到 50 个字符', trigger: 'blur' }
  ]
}

const filteredKnowledgeBases = computed(() => {
  if (!searchKeyword.value) return knowledgeBases.value
  const keyword = searchKeyword.value.toLowerCase()
  return knowledgeBases.value.filter(kb =>
    kb.name.toLowerCase().includes(keyword) ||
    kb.description?.toLowerCase().includes(keyword)
  )
})

const loadKnowledgeBases = async () => {
  listLoading.value = true
  try {
    const res = await knowledgeBaseAPI.getList({ page: page.value, page_size: pageSize.value })
    knowledgeBases.value = res.result || []
    total.value = res.total_count || 0
  } catch (error) {
    console.error('Failed to load knowledge bases:', error)
  } finally {
    listLoading.value = false
  }
}

const selectKnowledgeBase = (kb: KnowledgeBase) => {
  selectedKb.value = kb
}

const openEditDialog = () => {
  if (selectedKb.value) {
    editForm.name = selectedKb.value.name
    editForm.description = selectedKb.value.description || ''
    showEditDialog.value = true
  }
}

const handleCreate = async () => {
  if (!createFormRef.value) return
  await createFormRef.value.validate(async (valid) => {
    if (valid) {
      createLoading.value = true
      try {
        await knowledgeBaseAPI.create(createForm)
        ElMessage.success('创建成功')
        showCreateDialog.value = false
        createForm.name = ''
        createForm.description = ''
        loadKnowledgeBases()
      } catch (error) {
        console.error('Create failed:', error)
      } finally {
        createLoading.value = false
      }
    }
  })
}

const handleUpdate = async () => {
  if (!editFormRef.value || !selectedKb.value) return
  await editFormRef.value.validate(async (valid) => {
    if (valid) {
      editLoading.value = true
      try {
        const updated = await knowledgeBaseAPI.update(selectedKb.value!.id, editForm)
        ElMessage.success('更新成功')
        showEditDialog.value = false
        selectedKb.value = updated
        loadKnowledgeBases()
      } catch (error) {
        console.error('Update failed:', error)
      } finally {
        editLoading.value = false
      }
    }
  })
}

const handleDelete = async () => {
  if (!selectedKb.value) return
  try {
    await ElMessageBox.confirm(
      `确定要删除知识库"${selectedKb.value.name}"吗？此操作不可恢复。`,
      '警告',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await knowledgeBaseAPI.delete(selectedKb.value.id)
    ElMessage.success('删除成功')
    selectedKb.value = null
    loadKnowledgeBases()
  } catch (error) {
    if (error !== 'cancel') console.error('Delete failed:', error)
  }
}

const goToDocuments = () => {
  if (selectedKb.value) {
    router.push({ name: 'Document', params: { kbId: selectedKb.value.id } })
  }
}

const goToRetrieval = () => {
  if (selectedKb.value) {
    router.push({ name: 'Retrieval', params: { kbId: selectedKb.value.id } })
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadKnowledgeBases()
})
</script>

<style scoped>
.knowledge-base-page {
  min-height: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #1a1a2e;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.list-card, .detail-card {
  height: calc(100vh - 200px);
}

.kb-list {
  max-height: calc(100vh - 340px);
  overflow-y: auto;
}

.kb-item {
  display: flex;
  gap: 12px;
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.kb-item:hover {
  background: #f9fafb;
  border-color: #667eea;
}

.kb-item.active {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1), rgba(118, 75, 162, 0.1));
  border-color: #667eea;
}

.kb-icon {
  font-size: 32px;
}

.kb-info {
  flex: 1;
  overflow: hidden;
}

.kb-title {
  font-weight: 600;
  font-size: 16px;
  color: #1a1a2e;
  margin-bottom: 4px;
}

.kb-desc {
  font-size: 13px;
  color: #6b7280;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 4px;
}

.kb-meta {
  font-size: 12px;
  color: #9ca3af;
}

.pagination {
  margin-top: 16px;
  justify-content: center;
}

.action-buttons {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.empty-detail {
  height: calc(100vh - 200px);
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border-radius: 12px;
}
</style>
