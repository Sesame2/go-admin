package retriever

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/Sesame2/go-admin/internal/models"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// SearchOptions 定义检索选项
type SearchOptions struct {
	TopK       int     // 返回的最大结果数
	Threshold  float32 // 可选的相似度阈值（0-1之间）
	IncludeAll bool    // 是否包含所有元数据
}

// SearchResult 封装搜索结果和相似度分数
type SearchResult struct {
	Chunk *models.KnowledgeChunk // 知识块
	Score float32                // 相似度分数
}

// VectorRetriever 实现向量检索器
type VectorRetriever struct {
	db         *gorm.DB // GORM 数据库实例
	tableName  string   // 表名
	dimensions int      // 向量维度
	kbID       string   // 可选的知识库ID过滤
}

// NewVectorRetriever 创建一个新的向量检索器
func NewVectorRetriever(db *gorm.DB, options ...Option) *VectorRetriever {
	r := &VectorRetriever{
		db:         db,
		dimensions: 768,                // 默认向量维度为768
		tableName:  "knowledge_chunks", // 默认表名
	}

	// 应用选项
	for _, opt := range options {
		opt(r)
	}

	return r
}

// Option 定义检索器配置选项
type Option func(*VectorRetriever)

// WithDimensions 设置向量维度
func WithDimensions(dimensions int) Option {
	return func(r *VectorRetriever) {
		r.dimensions = dimensions
	}
}

// WithTableName 设置表名
func WithTableName(tableName string) Option {
	return func(r *VectorRetriever) {
		r.tableName = tableName
	}
}

// WithKnowledgeBase 设置默认知识库ID
func WithKnowledgeBase(kbID string) Option {
	return func(r *VectorRetriever) {
		r.kbID = kbID
	}
}

// Search 执行向量相似度搜索
func (r *VectorRetriever) Search(ctx context.Context, queryVector []float32, options SearchOptions) ([]SearchResult, error) {
	if len(queryVector) != r.dimensions {
		return nil, fmt.Errorf("query vector dimension mismatch: expected %d, got %d", r.dimensions, len(queryVector))
	}

	if options.TopK <= 0 {
		options.TopK = 5 // 默认返回5个结果
	}

	// 构建查询
	query := r.db.WithContext(ctx).Model(&models.KnowledgeChunk{})

	// 如果设置了知识库ID，添加过滤条件
	if r.kbID != "" {
		query = query.Where("kb_id = ?", r.kbID)
	}

	// 创建pgvector向量
	pgvectorQuery := pgvector.NewVector(queryVector)

	// 使用pgvector的余弦距离进行排序
	// 注意：需要在PostgreSQL中安装pgvector扩展
	query = query.Order(fmt.Sprintf("embedding <=> '%s'", pgvectorQuery.String()))

	// 限制结果数量
	query = query.Limit(options.TopK)

	// 执行查询
	var chunks []*models.KnowledgeChunk
	if err := query.Find(&chunks).Error; err != nil {
		return nil, fmt.Errorf("failed to execute vector search: %w", err)
	}

	// 组装结果
	results := make([]SearchResult, 0, len(chunks))
	for _, chunk := range chunks {
		// 计算向量相似度分数
		score := calculateSimilarity(queryVector, chunk.Embedding.Slice())

		// 应用阈值过滤（如果设置）
		if options.Threshold > 0 && score < options.Threshold {
			continue
		}

		results = append(results, SearchResult{
			Chunk: chunk,
			Score: score,
		})
	}

	return results, nil
}

// SearchByText 执行文本搜索，先将文本转换为向量再执行搜索
func (r *VectorRetriever) SearchByText(ctx context.Context, text string, options SearchOptions) ([]SearchResult, error) {
	// 将文本转换为向量（这里需要集成文本嵌入模型）
	vector, err := textToVector(text, r.dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to convert text to vector: %w", err)
	}

	return r.Search(ctx, vector, options)
}

// SearchByID 根据文档ID进行搜索
func (r *VectorRetriever) SearchByID(ctx context.Context, docID string, options SearchOptions) ([]SearchResult, error) {
	// 构建查询
	query := r.db.WithContext(ctx).Model(&models.KnowledgeChunk{}).Where("doc_id = ?", docID)

	// 如果设置了知识库ID，添加过滤条件
	if r.kbID != "" {
		query = query.Where("kb_id = ?", r.kbID)
	}

	// 执行查询
	var chunks []*models.KnowledgeChunk
	if err := query.Find(&chunks).Error; err != nil {
		return nil, fmt.Errorf("failed to execute search by doc ID: %w", err)
	}

	// 组装结果
	results := make([]SearchResult, 0, len(chunks))
	for _, chunk := range chunks {
		results = append(results, SearchResult{
			Chunk: chunk,
			Score: 1.0, // 完全匹配的文档ID，给予最高分数
		})
	}

	return results, nil
}

// SearchCombined 执行组合搜索（向量 + 文本过滤）
func (r *VectorRetriever) SearchCombined(
	ctx context.Context,
	queryVector []float32,
	textFilter string,
	kbID string,
	options SearchOptions,
) ([]SearchResult, error) {
	if len(queryVector) != r.dimensions {
		return nil, fmt.Errorf("query vector dimension mismatch: expected %d, got %d", r.dimensions, len(queryVector))
	}

	// 构建查询
	query := r.db.WithContext(ctx).Model(&models.KnowledgeChunk{})

	// 添加过滤条件
	if kbID != "" {
		query = query.Where("kb_id = ?", kbID)
	} else if r.kbID != "" {
		query = query.Where("kb_id = ?", r.kbID)
	}

	// 如果有文本过滤条件
	if textFilter != "" {
		query = query.Where("text LIKE ?", "%"+textFilter+"%")
	}

	// 创建pgvector向量
	pgvectorQuery := pgvector.NewVector(queryVector)

	// 使用pgvector的余弦距离进行排序
	query = query.Order(fmt.Sprintf("embedding <=> '%s'", pgvectorQuery.String()))

	// 限制结果数量
	query = query.Limit(options.TopK)

	// 执行查询
	var chunks []*models.KnowledgeChunk
	if err := query.Find(&chunks).Error; err != nil {
		return nil, fmt.Errorf("failed to execute combined search: %w", err)
	}

	// 组装结果
	results := make([]SearchResult, 0, len(chunks))
	for _, chunk := range chunks {
		// 计算向量相似度分数
		score := calculateSimilarity(queryVector, chunk.Embedding.Slice())

		// 应用阈值过滤（如果设置）
		if options.Threshold > 0 && score < options.Threshold {
			continue
		}

		results = append(results, SearchResult{
			Chunk: chunk,
			Score: score,
		})
	}

	return results, nil
}

// 辅助函数：计算向量相似度
func calculateSimilarity(vec1, vec2 []float32) float32 {
	// 实际实现应使用余弦相似度或欧氏距离等
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct float32
	var norm1 float32
	var norm2 float32

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	// 余弦相似度 = 点积 / (norm1 * norm2)的平方根
	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(norm1))) * float32(math.Sqrt(float64(norm2))))
}

// 辅助函数：文本转向量（需要集成文本嵌入模型）
func textToVector(text string, dimensions int) ([]float32, error) {
	// 这里应该集成文本嵌入模型
	// 例如调用OpenAI Embeddings API或本地嵌入模型

	// 占位实现，返回随机向量
	vector := make([]float32, dimensions)
	for i := 0; i < dimensions; i++ {
		vector[i] = rand.Float32()
	}

	// 规范化向量
	var norm float32
	for _, v := range vector {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))

	for i := range vector {
		vector[i] /= norm
	}

	return vector, nil
}
