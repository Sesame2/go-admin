<template>
  <div class="retrieval-page">
    <el-page-header @back="goBack">
      <template #content>
        <span class="page-title">💬 {{ kbName }} - 智能问答</span>
      </template>
    </el-page-header>

    <el-card class="chat-card" shadow="hover">
      <div class="chat-container">
        <!-- Messages Area -->
        <div ref="messagesContainer" class="messages-container">
          <el-empty v-if="conversations.length === 0" description="开始提问吧！" />

          <div v-for="(conv, idx) in conversations" :key="idx" class="conversation">
            <!-- User Message -->
            <div class="message user">
              <div class="message-content">
                <div class="message-bubble user-bubble">{{ conv.question }}</div>
                <el-avatar :size="36" class="avatar user-avatar">👤</el-avatar>
              </div>
            </div>

            <!-- AI Response -->
            <div class="message assistant">
              <div class="message-content">
                <el-avatar :size="36" class="avatar ai-avatar">🤖</el-avatar>
                <div class="message-bubble ai-bubble">
                  <!-- Status Bar -->
                  <div class="response-status">
                    <div :class="['status-dot', conv.status]"></div>
                    <span class="status-text">{{ conv.statusText }}</span>
                  </div>

                  <!-- Processing Steps -->
                  <div v-if="conv.steps.length > 0" class="steps-container">
                    <div
                      v-for="step in conv.steps"
                      :key="step.id"
                      :class="['step-item', { active: step.isLoading, completed: !step.isLoading }]"
                    >
                      <div class="step-header" @click="step.collapsed = !step.collapsed">
                        <div :class="['step-icon', step.isLoading ? 'loading' : step.type]">
                          {{ getStepIcon(step.type, step.isLoading) }}
                        </div>
                        <span class="step-title">{{ step.title }}</span>
                        <el-icon v-if="step.content" class="collapse-icon">
                          <component :is="step.collapsed ? 'ArrowRight' : 'ArrowDown'" />
                        </el-icon>
                      </div>
                      <div v-if="step.content && !step.collapsed" class="step-detail" v-html="step.content"></div>
                    </div>
                  </div>

                  <!-- Thinking Box -->
                  <div v-if="conv.thinking" class="thinking-box">
                    💭 {{ conv.thinking }}
                  </div>

                  <!-- Final Answer -->
                  <div v-if="conv.answer" class="answer-content" v-html="formatMarkdown(conv.answer)"></div>

                  <!-- Streaming indicator -->
                  <div v-if="conv.status === 'running' && !conv.answer" class="typing-indicator">
                    <span></span><span></span><span></span>
                  </div>

                  <!-- References -->
                  <div v-if="conv.references.length > 0" class="references-section">
                    <div class="references-header" @click="conv.refsCollapsed = !conv.refsCollapsed">
                      <span>📚 参考来源 ({{ conv.references.length }})</span>
                      <el-icon class="collapse-icon">
                        <component :is="conv.refsCollapsed ? 'ArrowRight' : 'ArrowDown'" />
                      </el-icon>
                    </div>
                    <div v-if="!conv.refsCollapsed" class="references-list">
                      <div v-for="(ref, i) in conv.references" :key="i" class="reference-item">
                        <div class="ref-title">📄 {{ ref.title || `来源 ${i + 1}` }}</div>
                        <div class="ref-content">{{ ref.content.substring(0, 150) }}...</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Input Area -->
        <div class="input-container">
          <el-input
            v-model="question"
            :disabled="isStreaming"
            placeholder="请输入您的问题..."
            size="large"
            @keyup.enter="handleSend"
          >
            <template #append>
              <el-button
                :loading="isStreaming"
                :disabled="!question.trim()"
                type="primary"
                @click="handleSend"
              >
                {{ isStreaming ? '发送中...' : '发送' }}
              </el-button>
            </template>
          </el-input>
          <div class="action-buttons">
            <el-button size="small" @click="handleClear">清空对话</el-button>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowRight, ArrowDown } from '@element-plus/icons-vue'
import { retrievalAPI, knowledgeBaseAPI } from '../api'

interface Step {
  id: string
  type: 'decompose' | 'retrieve' | 'select' | 'answer'
  title: string
  content: string
  isLoading: boolean
  collapsed: boolean
}

interface Reference {
  title: string
  content: string
}

interface Conversation {
  question: string
  status: 'running' | 'idle' | 'error'
  statusText: string
  steps: Step[]
  thinking: string
  answer: string
  references: Reference[]
  refsCollapsed: boolean
}

const route = useRoute()
const router = useRouter()

const kbId = ref(route.params.kbId as string)
const kbName = ref('')
const question = ref('')
const isStreaming = ref(false)
const conversations = ref<Conversation[]>([])
const messagesContainer = ref<HTMLElement | null>(null)

let currentConv: Conversation | null = null

const loadKnowledgeBase = async () => {
  try {
    const res = await knowledgeBaseAPI.getDetail(kbId.value)
    kbName.value = res.name
  } catch (error) {
    console.error('Failed to load KB:', error)
  }
}

const handleSend = async () => {
  if (!question.value.trim() || isStreaming.value) return

  // Create new conversation
  currentConv = reactive<Conversation>({
    question: question.value.trim(),
    status: 'running',
    statusText: '正在处理...',
    steps: [],
    thinking: '',
    answer: '',
    references: [],
    refsCollapsed: true
  })
  conversations.value.push(currentConv)
  question.value = ''
  isStreaming.value = true
  scrollToBottom()

  try {
    await retrievalAPI.streamRetrieve(
      kbId.value,
      currentConv.question,
      handleSSEEvent,
      (error) => {
        if (currentConv) {
          currentConv.status = 'error'
          currentConv.statusText = '连接错误'
        }
        ElMessage.error('检索失败，请重试')
      }
    )
  } catch (error) {
    console.error('Send failed:', error)
  } finally {
    isStreaming.value = false
    currentConv = null
  }
}

const handleSSEEvent = (eventType: string, data: any) => {
  if (!currentConv) return
  console.log('SSE:', eventType, data)

  switch (eventType) {
    case 'start':
      currentConv.statusText = data.message || '开始检索...'
      break

    case 'decompose_start':
      addStep(data.step_id, 'decompose', '🔍 问题分解', '', true)
      currentConv.statusText = data.message || '分解问题中...'
      break

    case 'decompose_result': {
      const d = data.data || {}
      let content = d.thinking ? `<div class="thinking-mini">💭 ${d.thinking}</div>` : ''
      if (d.sub_questions?.length > 0) {
        content += '<ul class="sub-list">' + d.sub_questions.map((q: string) => `<li>🔹 ${q}</li>`).join('') + '</ul>'
      } else {
        content += '<p class="no-sub">无需继续分解</p>'
      }
      updateStep(data.step_id, content, false)
      break
    }

    case 'retrieve_start':
      addStep(data.step_id + '_r', 'retrieve', '📥 信息检索', '', true)
      currentConv.statusText = data.message || '检索中...'
      break

    case 'retrieve_result': {
      const r = data.data || {}
      const candidates = r.atom_candidates || []
      let content = `<p>检索方式: <strong>${r.retrieval_method || 'unknown'}</strong> | 找到 ${candidates.length} 个候选</p>`
      if (candidates.length > 0) {
        content += '<div class="candidates">'
        candidates.slice(0, 3).forEach((c: any) => {
          content += `<div class="candidate">📌 ${c.atom_question || ''} <span class="score">${((c.retrieval_score || 0) * 100).toFixed(0)}%</span></div>`
        })
        if (candidates.length > 3) content += `<div class="more">...还有 ${candidates.length - 3} 个</div>`
        content += '</div>'
      }
      updateStep(data.step_id + '_r', content, false)
      break
    }

    case 'select_start':
      addStep(data.step_id + '_s', 'select', '✨ 信息选择', '', true)
      currentConv.statusText = data.message || '选择中...'
      break

    case 'select_result': {
      const s = data.data || {}
      let content = s.thinking ? `<div class="thinking-mini">💭 ${s.thinking}</div>` : ''
      if (s.selected && s.chosen_info) {
        content += `<div class="selected">✅ ${s.chosen_info.atom_question || ''}</div>`
      } else {
        content += '<p class="no-sub">❌ 未选择新信息</p>'
      }
      updateStep(data.step_id + '_s', content, false)
      break
    }

    case 'answer_start':
      addStep('answer', 'answer', '💬 生成答案', `基于 ${data.data?.references_count || 0} 条参考生成中...`, true)
      currentConv.statusText = data.message || '生成答案...'
      break

    case 'answer_chunk':
      currentConv.answer += data.data?.chunk || ''
      scrollToBottom()
      break

    case 'answer_complete': {
      const a = data.data || {}
      currentConv.answer = a.answer || currentConv.answer
      currentConv.thinking = a.thinking || ''
      if (a.references?.length > 0) {
        currentConv.references = a.references.map((r: any) => ({
          title: r.source_chunk_title || '',
          content: r.source_chunk_content || ''
        }))
      }
      updateStep('answer', '已完成 ✅', false)
      break
    }

    case 'complete':
      currentConv.status = 'idle'
      currentConv.statusText = `完成！共 ${data.data?.iterations || 0} 轮迭代`
      break

    case 'error':
      currentConv.status = 'error'
      currentConv.statusText = data.message || '发生错误'
      break

    case 'done':
      currentConv.status = 'idle'
      break
  }
  scrollToBottom()
}

const addStep = (id: string, type: Step['type'], title: string, content: string, isLoading: boolean) => {
  if (!currentConv) return
  currentConv.steps.push({ id, type, title, content, isLoading, collapsed: false })
}

const updateStep = (id: string, content: string, isLoading: boolean) => {
  if (!currentConv) return
  const step = currentConv.steps.find(s => s.id === id)
  if (step) {
    step.content = content
    step.isLoading = isLoading
  }
}

const getStepIcon = (type: string, isLoading: boolean) => {
  if (isLoading) return '⏳'
  const icons: Record<string, string> = { decompose: '✓', retrieve: '✓', select: '✓', answer: '✓' }
  return icons[type] || '•'
}

const formatMarkdown = (text: string) => {
  if (!text) return ''
  return text
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const handleClear = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有对话吗？', '提示', { type: 'warning' })
    conversations.value = []
    ElMessage.success('已清空')
  } catch {}
}

const goBack = () => router.push({ name: 'KnowledgeBase' })

onMounted(() => loadKnowledgeBase())
</script>

<style scoped>
.retrieval-page {
  min-height: 100%;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
}

.chat-card {
  margin-top: 20px;
  height: calc(100vh - 180px);
}

.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: linear-gradient(180deg, #f8f9fa 0%, #e9ecef 100%);
}

.conversation {
  margin-bottom: 24px;
}

.message {
  margin-bottom: 12px;
}

.message-content {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.message.user .message-content {
  justify-content: flex-end;
}

.avatar {
  flex-shrink: 0;
}

.user-avatar {
  background: linear-gradient(135deg, #667eea, #764ba2);
  order: 2;
}

.ai-avatar {
  background: #e3e8ee;
}

.message-bubble {
  max-width: 75%;
  padding: 14px 18px;
  border-radius: 18px;
  line-height: 1.6;
  animation: fadeIn 0.25s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.user-bubble {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  border-bottom-right-radius: 4px;
}

.ai-bubble {
  background: white;
  color: #333;
  border: 1px solid #e0e0e0;
  border-bottom-left-radius: 4px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}

/* Status */
.response-status {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.status-dot.running { background: #4caf50; animation: blink 1s infinite; }
.status-dot.idle { background: #9e9e9e; }
.status-dot.error { background: #f44336; }

@keyframes blink {
  50% { opacity: 0.3; }
}

.status-text {
  font-size: 0.85em;
  color: #666;
}

/* Steps */
.steps-container {
  margin-bottom: 14px;
}

.step-item {
  margin-bottom: 8px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e8e8e8;
}

.step-item.active { border-color: #667eea; }
.step-item.completed { border-color: #4caf50; }

.step-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  cursor: pointer;
  background: #fafafa;
  transition: background 0.2s;
}

.step-header:hover {
  background: #f0f0f0;
}

.step-icon {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: white;
}

.step-icon.decompose { background: #2196f3; }
.step-icon.retrieve { background: #ff9800; }
.step-icon.select { background: #9c27b0; }
.step-icon.answer { background: #4caf50; }
.step-icon.loading { background: #667eea; animation: pulse 1.2s infinite; }

@keyframes pulse {
  50% { opacity: 0.5; }
}

.step-title {
  flex: 1;
  font-weight: 500;
  font-size: 0.9em;
}

.collapse-icon {
  color: #999;
  transition: transform 0.2s;
}

.step-detail {
  padding: 10px 12px;
  background: #f9f9f9;
  font-size: 0.85em;
  color: #555;
  border-top: 1px solid #eee;
}

:deep(.thinking-mini) {
  background: #fff8e7;
  padding: 8px 10px;
  border-radius: 6px;
  margin-bottom: 8px;
  font-style: italic;
  color: #a67c00;
  border-left: 3px solid #ffc107;
}

:deep(.sub-list) {
  list-style: none;
  padding: 0;
  margin: 0;
}

:deep(.sub-list li) {
  padding: 4px 0;
  color: #444;
}

:deep(.no-sub) {
  color: #888;
  font-style: italic;
}

:deep(.candidates) {
  margin-top: 6px;
}

:deep(.candidate) {
  padding: 6px 8px;
  background: white;
  border-radius: 6px;
  margin-bottom: 4px;
  border: 1px solid #e0e0e0;
}

:deep(.candidate .score) {
  float: right;
  color: #888;
  font-size: 0.85em;
}

:deep(.more) {
  color: #888;
  font-size: 0.85em;
  text-align: right;
}

:deep(.selected) {
  background: #e8f5e9;
  padding: 8px 10px;
  border-radius: 6px;
  color: #2e7d32;
  border: 1px solid #a5d6a7;
}

/* Thinking */
.thinking-box {
  background: #fff8e1;
  border-left: 3px solid #ffc107;
  padding: 10px 12px;
  border-radius: 0 8px 8px 0;
  margin-bottom: 12px;
  font-style: italic;
  color: #6d5c00;
}

/* Answer */
.answer-content {
  font-size: 1em;
  line-height: 1.75;
  white-space: pre-wrap;
}

/* Typing indicator */
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 8px 0;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  background: #667eea;
  border-radius: 50%;
  animation: typing 1.4s infinite ease-in-out;
}

.typing-indicator span:nth-child(1) { animation-delay: 0s; }
.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typing {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.5; }
  40% { transform: scale(1); opacity: 1; }
}

/* References */
.references-section {
  margin-top: 14px;
  border-top: 1px solid #eee;
  padding-top: 12px;
}

.references-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  font-weight: 500;
  color: #555;
  padding: 6px 0;
}

.references-header:hover {
  color: #667eea;
}

.references-list {
  margin-top: 8px;
}

.reference-item {
  background: #f8f9fa;
  padding: 10px;
  border-radius: 8px;
  margin-bottom: 6px;
  border-left: 3px solid #667eea;
}

.ref-title {
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.ref-content {
  font-size: 0.85em;
  color: #666;
}

/* Input */
.input-container {
  padding: 16px 20px;
  background: white;
  border-top: 1px solid #eee;
}

.action-buttons {
  margin-top: 10px;
}

/* Scrollbar */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-track { background: #f1f1f1; border-radius: 3px; }
::-webkit-scrollbar-thumb { background: #ccc; border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: #aaa; }
</style>
