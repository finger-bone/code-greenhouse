package bootstrap

import (
	"judge/jConfig"
	"judge/schema"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func bootstrapSchema(logger *zap.Logger, db *gorm.DB) {
	err := db.AutoMigrate(
		&schema.User{},
		&schema.UserAttribute{},
		&schema.UserBasicAuthentication{},
		&schema.Repository{},
		&schema.Testing{},
		&schema.RepositoryTestingSerial{},
	)
	if err != nil {
		logger.Panic("Failed to migrate schema.")
	}
}

func bootstrapDatabase(logger *zap.Logger, config *jConfig.DatabaseConfig) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(config.DbFile), &gorm.Config{})
	if err != nil {
		logger.Panic("Failed to connect to db.")
	}
	bootstrapSchema(logger, db)
	return db
}
