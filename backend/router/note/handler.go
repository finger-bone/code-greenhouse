package note

import (
	"fmt"
	"judge/challenge"
	"judge/jConfig"
	"judge/router"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/russross/blackfriday"
	"go.uber.org/zap"
)

const (
	NOTE_FILE_TYPE_MARKDOWN = "markdown"
	NOTE_FILE_TYPE_WEBSITE  = "website"
)

func ServeWebSite(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	c *fiber.Ctx,
	challengeRecord *challenge.Challenge,
	stage int,
	theme string,
) error {
	requestPath := c.Params("*")
	noteFileOrPath := filepath.Join(
		config.Challenge.StorageFolder,
		challengeRecord.FolderName,
		challengeRecord.Stages[stage].NoteFileOrPath,
	)
	logger.Debug("Requesting", zap.String("path", requestPath))

	if requestPath == "" {
		return c.SendFile(noteFileOrPath + "/index.html")
	}

	fullPath := filepath.Join(noteFileOrPath, requestPath)
	return c.SendFile(fullPath)
}

func ServeMarkdown(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	c *fiber.Ctx,
	challengeRecord *challenge.Challenge,
	stage int,
	theme string,
) error {
	requestPath := c.Params("*")

	noteFileOrPath := filepath.Join(
		config.Challenge.StorageFolder,
		challengeRecord.FolderName,
		challengeRecord.Stages[stage].NoteFileOrPath,
	)

	if requestPath == "" {
		markdownFilePath := noteFileOrPath + "/index.md"
		markdownFile, err := os.ReadFile(markdownFilePath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(router.BuildError(
				"Internal Server Error",
			))
		}

		markdownStyleSheetPath := config.Challenge.MarkdownStyleSheetPath
		markdownStyleSheet, err := os.ReadFile(markdownStyleSheetPath)

		renderedMarkdown := blackfriday.MarkdownBasic(markdownFile)

		modeControl := `document.addEventListener("DOMContentLoaded", function () {
			const urlParams = new URLSearchParams(window.location.search);
			const theme = urlParams.get("theme");
			if (theme === "dark") {
				document.body.classList.add("dark");
			} else {
				document.body.classList.remove("dark");
			}
		});`

		template := `<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Note</title>
			<script>
				%s
			</script>
			<style>
			%s
			</style>
		</head>
		<body>
			%s
		</body>
		</html>`
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.Send([]byte(
			fmt.Sprintf(template, modeControl, string(markdownStyleSheet), string(renderedMarkdown)),
		))
	}

	fullPath := filepath.Join(noteFileOrPath, requestPath)
	return c.SendFile(fullPath)
}

func BuildNoteHandler(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		folderName := c.Params("folderName")
		stage, err := c.ParamsInt("stage", 0)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError(
				"Invalid stage",
			))
		}

		challengeRecord, err := challenge.ParseChallenge(
			logger,
			&config.Challenge,
			folderName,
		)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError(
				"Invalid challenge",
			))
		}

		if stage >= len(challengeRecord.Stages) || stage < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError(
				"Invalid stage",
			))
		}
		noteFileType := challengeRecord.Stages[stage].NoteFileType
		theme := c.Query("theme", "system")
		if noteFileType == NOTE_FILE_TYPE_MARKDOWN {
			return ServeMarkdown(
				logger,
				config,
				c,
				challengeRecord,
				stage,
				theme,
			)
		} else if noteFileType == NOTE_FILE_TYPE_WEBSITE {
			return ServeWebSite(
				logger,
				config,
				c,
				challengeRecord,
				stage,
				theme,
			)
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError(
				"Invalid note file type",
			))
		}
	}
}

func SetupNoteRouter(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	group *fiber.Router,
) {
	(*group).Get("/:folderName/:stage/*", BuildNoteHandler(logger, config))
}
