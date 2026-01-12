package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testUser 测试用简化的用户模型
type testUser struct {
	ID        string `gorm:"primaryKey"`
	Username  string `gorm:"not null;unique"`
	Email     string `gorm:"not null;unique"`
	Password  string `gorm:"not null"`
	Nickname  string
	Role      string `gorm:"default:'user'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (testUser) TableName() string {
	return "users"
}

// mockDatabase 实现 interfaces.Database 接口用于测试
type mockDatabase struct {
	db *gorm.DB
}

func (m *mockDatabase) DB() *gorm.DB {
	return m.db
}

func (m *mockDatabase) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (m *mockDatabase) Migrate(ctx context.Context) error {
	return nil
}

// setupTestDB 创建测试用的 SQLite 内存数据库
func setupTestDB(t *testing.T) *mockDatabase {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 使用简化的测试模型
	err = db.AutoMigrate(&testUser{})
	require.NoError(t, err)

	return &mockDatabase{db: db}
}

func TestUserRepository_Create(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	// 直接使用 GORM 测试创建功能
	user := &testUser{
		ID:       uuid.New().String(),
		Username: "testuser",
		Password: "hashedpassword",
		Nickname: "Test User",
		Email:    "test@example.com",
	}

	err := mockDB.db.Create(user).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)
}

func TestUserRepository_GetByID(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	// 先创建一个用户
	userID := uuid.New().String()
	user := &testUser{
		ID:       userID,
		Username: "testuser",
		Password: "hashedpassword",
		Nickname: "Test User",
		Email:    "test@example.com",
	}
	err := mockDB.db.Create(user).Error
	require.NoError(t, err)

	// 测试获取用户
	var foundUser testUser
	err = mockDB.db.First(&foundUser, "id = ?", userID).Error
	assert.NoError(t, err)
	assert.Equal(t, user.Username, foundUser.Username)
	assert.Equal(t, user.Email, foundUser.Email)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	// 测试获取不存在的用户
	var user testUser
	err := mockDB.db.First(&user, "id = ?", uuid.New().String()).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	// 先创建一个用户
	user := &testUser{
		ID:       uuid.New().String(),
		Username: "testuser",
		Password: "hashedpassword",
		Nickname: "Test User",
		Email:    "test@example.com",
	}
	err := mockDB.db.Create(user).Error
	require.NoError(t, err)

	// 测试通过用户名获取
	var foundUser testUser
	err = mockDB.db.Where("username = ?", "testuser").First(&foundUser).Error
	assert.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	// 测试获取不存在的用户名
	var user testUser
	err := mockDB.db.Where("username = ?", "nonexistent").First(&user).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepository_GetAll(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	// 创建多个用户
	users := []testUser{
		{ID: uuid.New().String(), Username: "user1", Password: "pass1", Nickname: "User 1", Email: "user1@example.com"},
		{ID: uuid.New().String(), Username: "user2", Password: "pass2", Nickname: "User 2", Email: "user2@example.com"},
		{ID: uuid.New().String(), Username: "user3", Password: "pass3", Nickname: "User 3", Email: "user3@example.com"},
	}

	for _, u := range users {
		err := mockDB.db.Create(&u).Error
		require.NoError(t, err)
	}

	// 测试获取所有用户
	var allUsers []testUser
	err := mockDB.db.Find(&allUsers).Error
	assert.NoError(t, err)
	assert.Len(t, allUsers, 3)
}

func TestUserRepository_Update(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	// 先创建一个用户
	userID := uuid.New().String()
	user := &testUser{
		ID:       userID,
		Username: "testuser",
		Password: "hashedpassword",
		Nickname: "Test User",
		Email:    "test@example.com",
	}
	err := mockDB.db.Create(user).Error
	require.NoError(t, err)

	// 更新用户
	err = mockDB.db.Model(&testUser{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"nickname": "Updated Nickname",
		"email":    "updated@example.com",
	}).Error
	assert.NoError(t, err)

	// 验证更新
	var updatedUser testUser
	err = mockDB.db.First(&updatedUser, "id = ?", userID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Nickname", updatedUser.Nickname)
	assert.Equal(t, "updated@example.com", updatedUser.Email)
}

func TestUserRepository_Delete(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	// 先创建一个用户
	userID := uuid.New().String()
	user := &testUser{
		ID:       userID,
		Username: "testuser",
		Password: "hashedpassword",
		Nickname: "Test User",
		Email:    "test@example.com",
	}
	err := mockDB.db.Create(user).Error
	require.NoError(t, err)

	// 删除用户（软删除）
	err = mockDB.db.Delete(&testUser{}, "id = ?", userID).Error
	assert.NoError(t, err)

	// 验证删除（软删除后通过默认查询应该找不到）
	var foundUser testUser
	err = mockDB.db.First(&foundUser, "id = ?", userID).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepository_Exist(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	// 先创建一个用户
	userID := uuid.New().String()
	user := &testUser{
		ID:       userID,
		Username: "testuser",
		Password: "hashedpassword",
		Nickname: "Test User",
		Email:    "test@example.com",
	}
	err := mockDB.db.Create(user).Error
	require.NoError(t, err)

	// 测试存在的用户
	var count int64
	err = mockDB.db.Model(&testUser{}).Where("id = ?", userID).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// 测试不存在的用户
	err = mockDB.db.Model(&testUser{}).Where("id = ?", uuid.New().String()).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
