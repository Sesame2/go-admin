import request from './request'
import type { Document, DocumentListResult, DownloadResponse } from '../types'

export const documentAPI = {
    // Get list of documents for a knowledge base
    getList(knowledgeBaseId: string, params: { page: number; page_size: number }): Promise<DocumentListResult> {
        return request.get('/documents', {
            params: {
                ...params,
                knowledge_base_id: knowledgeBaseId
            }
        })
    },

    getDetail(documentId: string): Promise<Document> {
        return request.get(`/documents/${documentId}`)
    },

    upload(knowledgeBaseId: string, file: File): Promise<Document> {
        const formData = new FormData()
        formData.append('knowledge_base_id', knowledgeBaseId)
        formData.append('file', file)
        return request.post('/documents/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        })
    },

    parse(documentId: string): Promise<{ message: string }> {
        return request.post('/documents/parse', { document_id: documentId })
    },

    getDownloadUrl(documentId: string): Promise<DownloadResponse> {
        return request.get(`/documents/${documentId}/download`)
    },

    delete(documentId: string): Promise<{ message: string }> {
        return request.delete(`/documents/${documentId}`)
    }
}
