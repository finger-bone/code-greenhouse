package bootstrap

import (
	"judge/jConfig"
	"judge/router/auth"
	"judge/router/note"
	"judge/router/query"
	"judge/router/repository"
	"judge/router/user"
	"judge/tester"

	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func bootstrapHandler(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	db *gorm.DB,
	docker *client.Client,
	app *fiber.App,
) {
	queryRouter := app.Group("/query")
	// POST /query graphql endpoint
	query.SetupQueryRouter(logger, config, db, &queryRouter)
	userRouter := app.Group("/user")
	// /user requires Bearer oauth token and Provider in header
	// GET /user/info endpoint, returns user info
	// GET /user/name endpoint, returns user git name
	// POST /user/password update password
	// query parameters: newPassword
	user.SetupUserRouter(logger, config, db, &userRouter)
	authRouter := app.Group("/auth")
	// GET /auth/single-user endpoint, returns if single user mode is enabled
	// GET /auth/url endpoint, returns auth url for oauth2
	// query parameters: provider, redirect_url, state
	// GET /auth/token endpoint, returns access token from oauth2
	// query parameters: provider, code
	// GET /auth/providers endpoint, returns all enabled auth providers
	// GET /auth/subject endpoint, returns subject from oauth2
	auth.SetupAuthRouter(logger, config, &authRouter)
	repoRouter := app.Group("/repo")
	// /repo requires Bearer oauth token and Provider in header
	// POST /repo/project create a new repo
	// query parameters: startpoint, folder
	// ALL /repo/git/{provider}/{subject}/{challengeFolderName}/{repoId} git server
	repository.SetupRepositoryRouter(logger, config, db, &repoRouter)
	testingRouter := app.Group("/testing")
	// /testing requires Bearer oauth token and Provider in header
	// POST /testing/pending push a new testing request
	// query repo, stage
	tester.SetupTestingRouter(logger, config, db, docker, &testingRouter)
	noteRouter := app.Group("/note")
	// /note
	// GET /note/:folderName/:stage/* Will return the note file, may be a markdown file or a webpage
	note.SetupNoteRouter(logger, config, &noteRouter)
}
