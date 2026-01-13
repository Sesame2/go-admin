export interface LoginInput {
    username: string
    password: string
}

export interface LoginResponse {
    token: string
}

export interface User {
    id: string
    username: string
    email: string
    role: string
    created_at: string
    updated_at: string
}

export interface KnowledgeBase {
    id: string
    name: string
    description: string
    user_id: string
    created_at: string
    updated_at: string
}

export interface KnowledgeBaseListResult {
    total_count: number
    total_pages: number
    current_page: number
    page_size: number
    result: KnowledgeBase[]
}

export interface CreateKnowledgeBaseInput {
    name: string
    description?: string
}

export interface UpdateKnowledgeBaseInput {
    name?: string
    description?: string
}

export interface Document {
    id: string
    knowledge_base_id: string
    filename: string
    file_type: string
    file_size: number | null
    status: DocumentStatus
    chunk_count: number
    error_message?: string
    created_at: string
    updated_at: string
}

export type DocumentStatus =
    | 'uploaded'
    | 'pending'
    | 'downloading'
    | 'parsing'
    | 'chunking'
    | 'embedding'
    | 'completed'
    | 'failed'

export interface DocumentListResult {
    documents: Document[]
    total: number
    page: number
    page_size: number
}

export interface ParseDocumentInput {
    document_id: string
}

export interface DownloadResponse {
    url: string
}

// SSE event types for streaming retrieval
export interface SSEEvent {
    type: 'event' | 'chunk' | 'done'
    event?: string
    message?: string
    content?: string
}

export interface ApiError {
    error: string
    detail?: string
    details?: string
}
