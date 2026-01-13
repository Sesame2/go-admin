package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// testKnowledgeBase 测试用简化的知识库模型
type testKnowledgeBase struct {
	ID            string `gorm:"primaryKey"`
	Name          string `gorm:"not null"`
	Description   string
	UserID        string `gorm:"not null;index"`
	DocumentCount int64  `gorm:"default:0"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (testKnowledgeBase) TableName() string {
	return "knowledge_bases"
}

// setupKnowledgeBaseTestDB 创建测试用的 SQLite 内存数据库
func setupKnowledgeBaseTestDB(t *testing.T) *mockDatabase {
	mockDB := setupTestDB(t)

	// 额外迁移知识库表
	err := mockDB.db.AutoMigrate(&testKnowledgeBase{})
	require.NoError(t, err)

	return mockDB
}

func TestKnowledgeBaseRepository_Create(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	kb := &testKnowledgeBase{
		ID:          uuid.New().String(),
		Name:        "Test KB",
		Description: "Test Knowledge Base",
		UserID:      uuid.New().String(),
	}

	err := mockDB.db.Create(kb).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, kb.ID)
}

func TestKnowledgeBaseRepository_GetByID(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	kbID := uuid.New().String()
	kb := &testKnowledgeBase{
		ID:          kbID,
		Name:        "Test KB",
		Description: "Test Knowledge Base",
		UserID:      uuid.New().String(),
	}
	err := mockDB.db.Create(kb).Error
	require.NoError(t, err)

	var foundKB testKnowledgeBase
	err = mockDB.db.First(&foundKB, "id = ?", kbID).Error
	assert.NoError(t, err)
	assert.Equal(t, kb.Name, foundKB.Name)
	assert.Equal(t, kb.Description, foundKB.Description)
}

func TestKnowledgeBaseRepository_GetByID_NotFound(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	var kb testKnowledgeBase
	err := mockDB.db.First(&kb, "id = ?", uuid.New().String()).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestKnowledgeBaseRepository_GetAll(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	userID := uuid.New().String()
	kbs := []testKnowledgeBase{
		{ID: uuid.New().String(), Name: "KB 1", Description: "Description 1", UserID: userID},
		{ID: uuid.New().String(), Name: "KB 2", Description: "Description 2", UserID: userID},
		{ID: uuid.New().String(), Name: "KB 3", Description: "Description 3", UserID: userID},
	}

	for _, kb := range kbs {
		err := mockDB.db.Create(&kb).Error
		require.NoError(t, err)
	}

	var allKBs []testKnowledgeBase
	err := mockDB.db.Limit(10).Offset(0).Find(&allKBs).Error
	assert.NoError(t, err)
	assert.Len(t, allKBs, 3)
}

func TestKnowledgeBaseRepository_GetAll_Pagination(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	userID := uuid.New().String()
	for i := 0; i < 5; i++ {
		kb := &testKnowledgeBase{
			ID:          uuid.New().String(),
			Name:        "KB",
			Description: "Description",
			UserID:      userID,
		}
		err := mockDB.db.Create(kb).Error
		require.NoError(t, err)
	}

	var page1 []testKnowledgeBase
	err := mockDB.db.Limit(2).Offset(0).Find(&page1).Error
	assert.NoError(t, err)
	assert.Len(t, page1, 2)

	var page2 []testKnowledgeBase
	err = mockDB.db.Limit(2).Offset(2).Find(&page2).Error
	assert.NoError(t, err)
	assert.Len(t, page2, 2)

	var page3 []testKnowledgeBase
	err = mockDB.db.Limit(2).Offset(4).Find(&page3).Error
	assert.NoError(t, err)
	assert.Len(t, page3, 1)
}

func TestKnowledgeBaseRepository_Count(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	userID := uuid.New().String()
	for i := 0; i < 3; i++ {
		kb := &testKnowledgeBase{
			ID:          uuid.New().String(),
			Name:        "KB",
			Description: "Description",
			UserID:      userID,
		}
		err := mockDB.db.Create(kb).Error
		require.NoError(t, err)
	}

	var count int64
	err := mockDB.db.Model(&testKnowledgeBase{}).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestKnowledgeBaseRepository_Update(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	kbID := uuid.New().String()
	kb := &testKnowledgeBase{
		ID:          kbID,
		Name:        "Test KB",
		Description: "Test Knowledge Base",
		UserID:      uuid.New().String(),
	}
	err := mockDB.db.Create(kb).Error
	require.NoError(t, err)

	kb.Name = "Updated KB"
	kb.Description = "Updated Description"
	err = mockDB.db.Save(kb).Error
	assert.NoError(t, err)

	var updatedKB testKnowledgeBase
	err = mockDB.db.First(&updatedKB, "id = ?", kbID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated KB", updatedKB.Name)
	assert.Equal(t, "Updated Description", updatedKB.Description)
}

func TestKnowledgeBaseRepository_Delete(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	kbID := uuid.New().String()
	kb := &testKnowledgeBase{
		ID:          kbID,
		Name:        "Test KB",
		Description: "Test Knowledge Base",
		UserID:      uuid.New().String(),
	}
	err := mockDB.db.Create(kb).Error
	require.NoError(t, err)

	err = mockDB.db.Delete(&testKnowledgeBase{}, "id = ?", kbID).Error
	assert.NoError(t, err)

	var foundKB testKnowledgeBase
	err = mockDB.db.First(&foundKB, "id = ?", kbID).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestKnowledgeBaseRepository_Exist(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	kbID := uuid.New().String()
	kb := &testKnowledgeBase{
		ID:          kbID,
		Name:        "Test KB",
		Description: "Test Knowledge Base",
		UserID:      uuid.New().String(),
	}
	err := mockDB.db.Create(kb).Error
	require.NoError(t, err)

	var count int64
	err = mockDB.db.Model(&testKnowledgeBase{}).Where("id = ?", kbID).Count(&count).Error
	assert.NoError(t, err)
	assert.True(t, count > 0)

	err = mockDB.db.Model(&testKnowledgeBase{}).Where("id = ?", uuid.New().String()).Count(&count).Error
	assert.NoError(t, err)
	assert.False(t, count > 0)
}

func TestKnowledgeBaseRepository_IncrementDocumentCount(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	kbID := uuid.New().String()
	kb := &testKnowledgeBase{
		ID:            kbID,
		Name:          "Test KB",
		Description:   "Test Knowledge Base",
		UserID:        uuid.New().String(),
		DocumentCount: 0,
	}
	err := mockDB.db.Create(kb).Error
	require.NoError(t, err)

	err = mockDB.db.Model(&testKnowledgeBase{}).Where("id = ?", kbID).
		UpdateColumn("document_count", gorm.Expr("document_count + ?", 1)).Error
	assert.NoError(t, err)

	var updatedKB testKnowledgeBase
	err = mockDB.db.First(&updatedKB, "id = ?", kbID).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), updatedKB.DocumentCount)

	err = mockDB.db.Model(&testKnowledgeBase{}).Where("id = ?", kbID).
		UpdateColumn("document_count", gorm.Expr("document_count + ?", 1)).Error
	assert.NoError(t, err)

	err = mockDB.db.First(&updatedKB, "id = ?", kbID).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(2), updatedKB.DocumentCount)
}

func TestKnowledgeBaseRepository_DecrementDocumentCount(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	kbID := uuid.New().String()
	kb := &testKnowledgeBase{
		ID:            kbID,
		Name:          "Test KB",
		Description:   "Test Knowledge Base",
		UserID:        uuid.New().String(),
		DocumentCount: 5,
	}
	err := mockDB.db.Create(kb).Error
	require.NoError(t, err)

	err = mockDB.db.Model(&testKnowledgeBase{}).Where("id = ? AND document_count > 0", kbID).
		UpdateColumn("document_count", gorm.Expr("document_count - ?", 1)).Error
	assert.NoError(t, err)

	var updatedKB testKnowledgeBase
	err = mockDB.db.First(&updatedKB, "id = ?", kbID).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(4), updatedKB.DocumentCount)
}

func TestKnowledgeBaseRepository_WithTransaction(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	err := mockDB.db.Transaction(func(tx *gorm.DB) error {
		kb := &testKnowledgeBase{
			ID:          uuid.New().String(),
			Name:        "TX KB",
			Description: "Transaction Test",
			UserID:      uuid.New().String(),
		}
		return tx.Create(kb).Error
	})
	assert.NoError(t, err)

	var kbs []testKnowledgeBase
	err = mockDB.db.Find(&kbs).Error
	assert.NoError(t, err)
	assert.Len(t, kbs, 1)
}
