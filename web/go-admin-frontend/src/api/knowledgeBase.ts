import request from './request'
import type {
    KnowledgeBase,
    KnowledgeBaseListResult,
    CreateKnowledgeBaseInput,
    UpdateKnowledgeBaseInput
} from '../types'

export const knowledgeBaseAPI = {
    // Get list of knowledge bases
    getList(params: { page: number; page_size: number }): Promise<KnowledgeBaseListResult> {
        return request.get('/knowledge_bases', { params })
    },

    getDetail(id: string): Promise<KnowledgeBase> {
        return request.get(`/knowledge_bases/${id}`)
    },

    create(data: CreateKnowledgeBaseInput): Promise<KnowledgeBase> {
        return request.post('/knowledge_bases', data)
    },

    update(id: string, data: UpdateKnowledgeBaseInput): Promise<KnowledgeBase> {
        return request.put(`/knowledge_bases/${id}`, data)
    },

    delete(id: string): Promise<{ message: string }> {
        return request.delete(`/knowledge_bases/${id}`)
    }
}
