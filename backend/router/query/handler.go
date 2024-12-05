package query

import (
	"judge/jConfig"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func getSchema() string {
	return `
	type StartPoint {
		name: String!
		description: [String!]!
	}
	
	type Stage {
		name: String!
		description: [String!]!
		noteFileOrPath: String!
		noteFileType: String!
	}

	type Challenge {
		folderName: String!
		startPoints: [StartPoint!]!
		stages: [Stage!]!
		basic: Basic!
	}
	
	type Basic {
		author: String!
		source: String!
		title: String!
		description: [String!]!
	}
	
	type User {
		subject: String!
		provider: String!
		createTime: String!
		updateTime: String!
		attributes: [UserAttribute!]!
	}
	
	type UserAttribute {
		key: String!
		value: String!
	}

	type Testing {
		repositoryId: String!
		serial: Int!
		stage: Int!
		status: String!
		message: String!
		log: String!
		createTime: String!
		runStartTime: String!
		runEndTime: String!
	}
	
	type Repository {
		repositoryId: String!
		subject: String!
		provider: String!
		challengeFolderName: String!
		startpoint: String!
		stage: Int!
		totalStages: Int!
		createTime: String!
		updateTime: String!
	}
	
	type Query {
		challenge(folderName: String!): Challenge
		challenges: [Challenge!]!
		repositories(subject: String!, provider: String!): [Repository!]!
		repository(repositoryId: String!): Repository
		user(subject: String!, provider: String!): User
		testingsByRepository(repositoryId: String!): [Testing!]!
		testing(repositoryId: String!, serial: Int!): Testing
		testingsByStage(repositoryId: String!, stage: Int!): [Testing!]!
	}
	`
}

type r struct {
	logger *zap.Logger
	config *jConfig.JudgeConfig
	db     *gorm.DB
}

func SetupQueryRouter(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	db *gorm.DB,
	group *fiber.Router,
) {
	s := getSchema()
	schema := graphql.MustParseSchema(s, &r{
		logger: logger,
		config: config,
		db:     db,
	}, graphql.UseFieldResolvers())
	handler := &relay.Handler{Schema: schema}
	(*group).Post("/", adaptor.HTTPHandlerFunc(handler.ServeHTTP))
}
