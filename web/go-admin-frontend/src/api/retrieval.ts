import type { SSEEvent } from '../types'

const API_BASE = 'http://localhost:8080/api'

export const retrievalAPI = {
    async streamRetrieve(
        knowledgeBaseId: string,
        question: string,
        onEvent: (eventType: string, data: any) => void,
        onError?: (error: Error) => void
    ): Promise<void> {
        const token = localStorage.getItem('token')
        const url = `${API_BASE}/retrieval/stream?knowledge_base_id=${encodeURIComponent(knowledgeBaseId)}&question=${encodeURIComponent(question)}`

        try {
            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'Authorization': token ? `Bearer ${token}` : '',
                    'Accept': 'text/event-stream'
                }
            })

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`)
            }

            const reader = response.body?.getReader()
            if (!reader) {
                throw new Error('No response body')
            }

            const decoder = new TextDecoder()
            let buffer = ''
            let currentEvent = ''

            while (true) {
                const { done, value } = await reader.read()
                if (done) break

                buffer += decoder.decode(value, { stream: true })
                const lines = buffer.split('\n')
                buffer = lines.pop() || ''

                for (const line of lines) {
                    if (line.startsWith('event: ')) {
                        // 保存事件类型
                        currentEvent = line.slice(7).trim()
                    } else if (line.startsWith('data: ')) {
                        const dataStr = line.slice(6).trim()
                        if (dataStr === '[DONE]') {
                            onEvent('done', {})
                            return
                        }
                        try {
                            const data = JSON.parse(dataStr)
                            // 如果没有显式的 event 行，尝试从 data 中获取
                            const eventType = currentEvent || data.event || 'unknown'
                            onEvent(eventType, data)
                            currentEvent = '' // 重置
                        } catch (e) {
                            console.error('Failed to parse SSE data:', e, dataStr)
                        }
                    }
                }
            }
        } catch (error) {
            console.error('SSE connection error:', error)
            if (onError) onError(error as Error)
        }
    }
}
