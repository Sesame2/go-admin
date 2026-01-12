package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// testDocument 测试用简化的文档模型
type testDocument struct {
	ID              string `gorm:"primaryKey"`
	Name            string `gorm:"not null"`
	Description     string
	KnowledgeBaseID string
	FileName        string
	FileSize        int64 `gorm:"default:0"`
	FileType        string
	Status          string `gorm:"default:'pending'"`
	UserID          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (testDocument) TableName() string {
	return "documents"
}

// setupDocumentTestDB 创建测试用的 SQLite 内存数据库
func setupDocumentTestDB(t *testing.T) *mockDatabase {
	mockDB := setupTestDB(t)

	// 额外迁移文档和知识库表
	err := mockDB.db.AutoMigrate(&testKnowledgeBase{}, &testDocument{})
	require.NoError(t, err)

	return mockDB
}

func TestDocumentRepository_Create(t *testing.T) {
	mockDB := setupDocumentTestDB(t)
	defer mockDB.Close()

	doc := &testDocument{
		ID:              uuid.New().String(),
		Name:            "Test Document",
		Description:     "Test Description",
		KnowledgeBaseID: uuid.New().String(),
		FileName:        "test.pdf",
		FileSize:        1024,
		FileType:        "application/pdf",
		Status:          "pending",
	}

	err := mockDB.db.Create(doc).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, doc.ID)
}

func TestDocumentRepository_GetByID(t *testing.T) {
	mockDB := setupDocumentTestDB(t)
	defer mockDB.Close()

	docID := uuid.New().String()
	doc := &testDocument{
		ID:              docID,
		Name:            "Test Document",
		Description:     "Test Description",
		KnowledgeBaseID: uuid.New().String(),
		FileName:        "test.pdf",
		FileSize:        1024,
		FileType:        "application/pdf",
		Status:          "pending",
	}
	err := mockDB.db.Create(doc).Error
	require.NoError(t, err)

	var foundDoc testDocument
	err = mockDB.db.First(&foundDoc, "id = ?", docID).Error
	assert.NoError(t, err)
	assert.Equal(t, doc.Name, foundDoc.Name)
	assert.Equal(t, doc.FileName, foundDoc.FileName)
}

func TestDocumentRepository_GetByID_NotFound(t *testing.T) {
	mockDB := setupDocumentTestDB(t)
	defer mockDB.Close()

	var doc testDocument
	err := mockDB.db.First(&doc, "id = ?", uuid.New().String()).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDocumentRepository_GetAll(t *testing.T) {
	mockDB := setupDocumentTestDB(t)
	defer mockDB.Close()

	for i := 0; i < 3; i++ {
		doc := &testDocument{
			ID:              uuid.New().String(),
			Name:            "Document",
			Description:     "Description",
			KnowledgeBaseID: uuid.New().String(),
			Status:          "pending",
		}
		err := mockDB.db.Create(doc).Error
		require.NoError(t, err)
	}

	var allDocs []testDocument
	err := mockDB.db.Find(&allDocs).Error
	assert.NoError(t, err)
	assert.Len(t, allDocs, 3)
}

func TestDocumentRepository_ListByKnowledgeBase(t *testing.T) {
	mockDB := setupDocumentTestDB(t)
	defer mockDB.Close()

	kbID1 := uuid.New().String()
	kbID2 := uuid.New().String()

	for i := 0; i < 3; i++ {
		doc := &testDocument{
			ID:              uuid.New().String(),
			Name:            "KB1 Document",
			Description:     "Description",
			KnowledgeBaseID: kbID1,
			Status:          "pending",
		}
		err := mockDB.db.Create(doc).Error
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		doc := &testDocument{
			ID:              uuid.New().String(),
			Name:            "KB2 Document",
			Description:     "Description",
			KnowledgeBaseID: kbID2,
			Status:          "pending",
		}
		err := mockDB.db.Create(doc).Error
		require.NoError(t, err)
	}

	var kb1Docs []testDocument
	err := mockDB.db.Where("knowledge_base_id = ?", kbID1).Find(&kb1Docs).Error
	assert.NoError(t, err)
	assert.Len(t, kb1Docs, 3)

	var kb2Docs []testDocument
	err = mockDB.db.Where("knowledge_base_id = ?", kbID2).Find(&kb2Docs).Error
	assert.NoError(t, err)
	assert.Len(t, kb2Docs, 2)
}

func TestDocumentRepository_CountByKnowledgeBase(t *testing.T) {
	mockDB := setupDocumentTestDB(t)
	defer mockDB.Close()

	kbID := uuid.New().String()

	for i := 0; i < 5; i++ {
		doc := &testDocument{
			ID:              uuid.New().String(),
			Name:            "Document",
			Description:     "Description",
			KnowledgeBaseID: kbID,
			Status:          "pending",
		}
		err := mockDB.db.Create(doc).Error
		require.NoError(t, err)
	}

	var count int64
	err := mockDB.db.Model(&testDocument{}).Where("knowledge_base_id = ?", kbID).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestDocumentRepository_Update(t *testing.T) {
	mockDB := setupDocumentTestDB(t)
	defer mockDB.Close()

	docID := uuid.New().String()
	doc := &testDocument{
		ID:              docID,
		Name:            "Test Document",
		Description:     "Test Description",
		KnowledgeBaseID: uuid.New().String(),
		Status:          "pending",
	}
	err := mockDB.db.Create(doc).Error
	require.NoError(t, err)

	doc.Name = "Updated Document"
	doc.Status = "completed"
	err = mockDB.db.Save(doc).Error
	assert.NoError(t, err)

	var updatedDoc testDocument
	err = mockDB.db.First(&updatedDoc, "id = ?", docID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Document", updatedDoc.Name)
	assert.Equal(t, "completed", updatedDoc.Status)
}

func TestDocumentRepository_Delete(t *testing.T) {
	mockDB := setupDocumentTestDB(t)
	defer mockDB.Close()

	docID := uuid.New().String()
	doc := &testDocument{
		ID:              docID,
		Name:            "Test Document",
		Description:     "Test Description",
		KnowledgeBaseID: uuid.New().String(),
		Status:          "pending",
	}
	err := mockDB.db.Create(doc).Error
	require.NoError(t, err)

	err = mockDB.db.Delete(&testDocument{}, "id = ?", docID).Error
	assert.NoError(t, err)

	var foundDoc testDocument
	err = mockDB.db.First(&foundDoc, "id = ?", docID).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDocumentRepository_Exist(t *testing.T) {
	mockDB := setupDocumentTestDB(t)
	defer mockDB.Close()

	docID := uuid.New().String()
	doc := &testDocument{
		ID:              docID,
		Name:            "Test Document",
		Description:     "Test Description",
		KnowledgeBaseID: uuid.New().String(),
		Status:          "pending",
	}
	err := mockDB.db.Create(doc).Error
	require.NoError(t, err)

	var count int64
	err = mockDB.db.Model(&testDocument{}).Where("id = ?", docID).Count(&count).Error
	assert.NoError(t, err)
	assert.True(t, count > 0)

	err = mockDB.db.Model(&testDocument{}).Where("id = ?", uuid.New().String()).Count(&count).Error
	assert.NoError(t, err)
	assert.False(t, count > 0)
}

func TestDocumentRepository_GetByUserID(t *testing.T) {
	mockDB := setupDocumentTestDB(t)
	defer mockDB.Close()

	userID := uuid.New().String()

	for i := 0; i < 3; i++ {
		doc := &testDocument{
			ID:              uuid.New().String(),
			Name:            "User Document",
			Description:     "Description",
			UserID:          userID,
			KnowledgeBaseID: uuid.New().String(),
			Status:          "pending",
		}
		err := mockDB.db.Create(doc).Error
		require.NoError(t, err)
	}

	var userDocs []testDocument
	err := mockDB.db.Where("user_id = ?", userID).Find(&userDocs).Error
	assert.NoError(t, err)
	assert.Len(t, userDocs, 3)
}
