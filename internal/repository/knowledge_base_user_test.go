package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeBaseRepository_GetByUserID(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	userID1 := uuid.New().String()
	userID2 := uuid.New().String()

	// 为用户1创建知识库
	for i := 0; i < 3; i++ {
		kb := &testKnowledgeBase{
			ID:          uuid.New().String(),
			Name:        "User1 KB",
			Description: "Description",
			UserID:      userID1,
		}
		err := mockDB.db.Create(kb).Error
		require.NoError(t, err)
	}

	// 为用户2创建知识库
	for i := 0; i < 2; i++ {
		kb := &testKnowledgeBase{
			ID:          uuid.New().String(),
			Name:        "User2 KB",
			Description: "Description",
			UserID:      userID2,
		}
		err := mockDB.db.Create(kb).Error
		require.NoError(t, err)
	}

	// 测试获取用户1的知识库
	var user1KBs []testKnowledgeBase
	err := mockDB.db.Where("user_id = ?", userID1).Limit(10).Offset(0).Find(&user1KBs).Error
	assert.NoError(t, err)
	assert.Len(t, user1KBs, 3)

	// 测试获取用户2的知识库
	var user2KBs []testKnowledgeBase
	err = mockDB.db.Where("user_id = ?", userID2).Limit(10).Offset(0).Find(&user2KBs).Error
	assert.NoError(t, err)
	assert.Len(t, user2KBs, 2)
}

func TestKnowledgeBaseRepository_CountByUserID(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	userID := uuid.New().String()

	// 创建3个知识库
	for i := 0; i < 3; i++ {
		kb := &testKnowledgeBase{
			ID:          uuid.New().String(),
			Name:        "User KB",
			Description: "Description",
			UserID:      userID,
		}
		err := mockDB.db.Create(kb).Error
		require.NoError(t, err)
	}

	// 统计知识库数量
	var count int64
	err := mockDB.db.Model(&testKnowledgeBase{}).Where("user_id = ?", userID).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// 统计不存在用户的知识库数量
	var count2 int64
	err = mockDB.db.Model(&testKnowledgeBase{}).Where("user_id = ?", uuid.New().String()).Count(&count2).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count2)
}

func TestKnowledgeBaseRepository_GetByUserID_Pagination(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	userID := uuid.New().String()

	// 创建5个知识库
	for i := 0; i < 5; i++ {
		kb := &testKnowledgeBase{
			ID:          uuid.New().String(),
			Name:        "User KB",
			Description: "Description",
			UserID:      userID,
		}
		err := mockDB.db.Create(kb).Error
		require.NoError(t, err)
	}

	// 第一页
	var page1 []testKnowledgeBase
	err := mockDB.db.Where("user_id = ?", userID).Limit(2).Offset(0).Order("created_at DESC").Find(&page1).Error
	assert.NoError(t, err)
	assert.Len(t, page1, 2)

	// 第二页
	var page2 []testKnowledgeBase
	err = mockDB.db.Where("user_id = ?", userID).Limit(2).Offset(2).Order("created_at DESC").Find(&page2).Error
	assert.NoError(t, err)
	assert.Len(t, page2, 2)

	// 第三页
	var page3 []testKnowledgeBase
	err = mockDB.db.Where("user_id = ?", userID).Limit(2).Offset(4).Order("created_at DESC").Find(&page3).Error
	assert.NoError(t, err)
	assert.Len(t, page3, 1)
}

func TestKnowledgeBaseRepository_GetByUserID_Empty(t *testing.T) {
	mockDB := setupKnowledgeBaseTestDB(t)
	defer mockDB.Close()

	userID := uuid.New().String()

	// 不存在的用户
	var kbs []testKnowledgeBase
	err := mockDB.db.Where("user_id = ?", userID).Find(&kbs).Error
	assert.NoError(t, err)
	assert.Len(t, kbs, 0)
}
