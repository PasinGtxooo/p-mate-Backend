package routes

import (
	"os"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/groq"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/middlewares"

	chatUC "github.com/FrostBitzX/smart-task-ai/internal/application/chat/usecase"
	dashboardUseCase "github.com/FrostBitzX/smart-task-ai/internal/application/dashboard/usecase"
	"github.com/FrostBitzX/smart-task-ai/internal/application/file"
	invitationUC "github.com/FrostBitzX/smart-task-ai/internal/application/invitation/usecase"
	profileUC "github.com/FrostBitzX/smart-task-ai/internal/application/profile/usecase"
	projectUC "github.com/FrostBitzX/smart-task-ai/internal/application/project/usecase"
	"github.com/FrostBitzX/smart-task-ai/internal/application/suggestion"
	suggestionUC "github.com/FrostBitzX/smart-task-ai/internal/application/suggestion/usecase"
	taskUC "github.com/FrostBitzX/smart-task-ai/internal/application/task/usecase"
	chatDomain "github.com/FrostBitzX/smart-task-ai/internal/domain/chats/service"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles"
	profileDomain "github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/service"
	projectDomain "github.com/FrostBitzX/smart-task-ai/internal/domain/projects/service"
	taskDomain "github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/service"
	repo "github.com/FrostBitzX/smart-task-ai/internal/infrastructure/persistence"
	handler "github.com/FrostBitzX/smart-task-ai/internal/infrastructure/rest"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterPrivateRoutes(app fiber.Router, db *gorm.DB, log logger.Logger) {
	api := app.Group("/api", middlewares.JWTMiddleware())

	// Profile setup
	profileRepository := repo.NewProfileRepository(db)
	accountRepository := repo.NewAccountRepository(db)
	profileService := profileDomain.NewProfileService(profileRepository, accountRepository)
	createProfileUC := profileUC.NewCreateProfileUseCase(profileService, log)
	getProfileUC := profileUC.NewGetProfileUseCase(profileService, log)
	updateProfileUC := profileUC.NewUpdateProfileUseCase(profileService, log)
	profileHandlerInstance := handler.NewProfileHandler(createProfileUC, getProfileUC, updateProfileUC, log)

	// Profile routes
	api.Post("/profiles", profileHandlerInstance.CreateProfile)
	api.Get("/profiles", profileHandlerInstance.GetProfile)
	api.Patch("/profiles", profileHandlerInstance.UpdateProfile)

	// Project setup
	projectRepository := repo.NewProjectRepository(db)
	taskRepository := repo.NewTaskRepository(db)
	memberRepository := repo.NewMemberRepository(db)
	projectService := projectDomain.NewProjectService(projectRepository, taskRepository, memberRepository)
	createProjectUC := projectUC.NewCreateProjectUseCase(projectService, log)
	listProjectByAccountUC := projectUC.NewListProjectByAccountUseCase(projectService, log)
	getProjectByIDUC := projectUC.NewGetProjectByIDUseCase(projectService, log)
	updateProjectUC := projectUC.NewUpdateProjectUseCase(projectService, log)
	deleteProjectUC := projectUC.NewDeleteProjectUseCase(projectService, log)
	addMemberUC := projectUC.NewAddMemberUseCase(projectService, accountRepository, log)
	removeMemberUC := projectUC.NewRemoveMemberUseCase(projectService, log)
	listMembersUC := projectUC.NewListProjectMembersUseCase(projectService, log)
	projectHandlerInstance := handler.NewProjectHandler(
		createProjectUC,
		listProjectByAccountUC,
		getProjectByIDUC,
		updateProjectUC,
		deleteProjectUC,
		addMemberUC,
		removeMemberUC,
		listMembersUC,
		log,
	)

	// Project routes
	api.Post("/projects", projectHandlerInstance.CreateProject)
	api.Get("/projects", projectHandlerInstance.ListProject)
	api.Get("/projects/:projectId", projectHandlerInstance.GetProject)
	api.Patch("/projects/:projectId", projectHandlerInstance.UpdateProject)
	api.Delete("/projects/:projectId", projectHandlerInstance.DeleteProject)
	api.Post("/projects/:projectId/members", projectHandlerInstance.AddMember)
	api.Delete("/projects/:projectId/members/:accountId", projectHandlerInstance.RemoveMember)
	api.Get("/projects/:projectId/members", projectHandlerInstance.ListMembers)

	// Invitation setup
	invitationService := projectDomain.NewInvitationService(
		repo.NewInvitationRepository(db),
		accountRepository,
		repo.NewMemberRepository(db),
		projectRepository,
	)
	createInvitationUC := invitationUC.NewCreateInvitationUseCase(invitationService, log)
	acceptInvitationUC := invitationUC.NewAcceptInvitationUseCase(invitationService, log)
	rejectInvitationUC := invitationUC.NewRejectInvitationUseCase(invitationService, log)
	cancelInvitationUC := invitationUC.NewCancelInvitationUseCase(invitationService, log)
	listMyInvitationsUC := invitationUC.NewListMyInvitationsUseCase(invitationService, log)
	listProjectInvitationsUC := invitationUC.NewListProjectInvitationsUseCase(invitationService, log)
	invitationHandlerInstance := handler.NewInvitationHandler(
		createInvitationUC,
		acceptInvitationUC,
		rejectInvitationUC,
		cancelInvitationUC,
		listMyInvitationsUC,
		listProjectInvitationsUC,
		log,
	)

	// Invitation routes
	api.Post("/projects/:projectId/invite", invitationHandlerInstance.CreateInvitation)
	api.Post("/projects/:projectId/invite/:inviteeAccountId/accept", invitationHandlerInstance.AcceptInvitation)
	api.Post("/projects/:projectId/invite/:inviteeAccountId/reject", invitationHandlerInstance.RejectInvitation)
	api.Delete("/projects/:projectId/invite/:inviteeAccountId", invitationHandlerInstance.CancelInvitation)
	api.Get("/accounts/me/invite", invitationHandlerInstance.ListMyInvitations)
	api.Get("/projects/:projectId/invite", invitationHandlerInstance.ListProjectInvitations)

	// Task setup
	taskService := taskDomain.NewTaskService(taskRepository, projectRepository, memberRepository)
	createTaskUC := taskUC.NewCreateTaskUseCase(taskService, log)
	getTaskByIDUC := taskUC.NewGetTaskByIDUseCase(taskService, log)
	listTasksByProjectUC := taskUC.NewListTasksByProjectUseCase(taskService, log)
	updateTaskUC := taskUC.NewUpdateTaskUseCase(taskService, log)
	deleteTaskUC := taskUC.NewDeleteTaskUseCase(taskService, log)
	taskHandlerInstance := handler.NewTaskHandler(createTaskUC, getTaskByIDUC, listTasksByProjectUC, updateTaskUC, deleteTaskUC, log)

	// Task routes
	api.Post("/:projectId/tasks", taskHandlerInstance.CreateTask)
	api.Get("/:projectId/tasks", taskHandlerInstance.ListTasksByProject)
	api.Get("/tasks/:taskId", taskHandlerInstance.GetTaskByID)
	api.Patch("/tasks/:taskId", taskHandlerInstance.UpdateTask)
	api.Delete("/tasks/:taskId", taskHandlerInstance.DeleteTask)

	// Chat setup
	groqClient, err := groq.NewGroqClient()
	if err != nil {
		log.Warn("Failed to initialize Groq client, chat endpoints will not be available", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		chatService := chatDomain.NewChatService(groqClient, taskService, projectService)
		sendMessageUC := chatUC.NewSendMessageUseCase(chatService, log)
		chatHandlerInstance := handler.NewChatHandler(sendMessageUC, log)

		// Chat routes (protected by JWT middleware via /api group)
		api.Post("/:projectId/chat", chatHandlerInstance.SendMessage)
	}

	// Dashboard setup
	getTaskStatisticsUC := dashboardUseCase.NewGetTaskStatisticsUseCase(taskService, log)
	getUnscheduledTasksUC := dashboardUseCase.NewGetUnscheduledTasksUseCase(taskService, log)
	listTodayTasksUC := dashboardUseCase.NewListTodayTasksUseCase(taskService, log)
	dashboardHandlerInstance := handler.NewDashboardHandler(getTaskStatisticsUC, getUnscheduledTasksUC, listTodayTasksUC, log)

	// Dashboard routes
	dashboard := api.Group("/dashboard")
	dashboard.Get("/statistics", dashboardHandlerInstance.GetTaskStatistics)
	dashboard.Get("/unscheduled-tasks", dashboardHandlerInstance.GetUnscheduledTasks)
	dashboard.Get("/today-tasks", dashboardHandlerInstance.ListTodayTasks)

	// Suggestion setup (manday-based teammate matching + manpower)
	suggestionService := suggestion.NewSuggestionService(taskRepository, projectRepository, memberRepository, profileRepository, accountRepository)
	suggestTeammatesUC := suggestionUC.NewSuggestTeammatesUseCase(suggestionService, log)
	getManpowerUC := suggestionUC.NewGetManpowerUseCase(suggestionService, log)
	setMemberSkillsUC := suggestionUC.NewSetMemberSkillsUseCase(suggestionService, log)
	listMemberSkillsUC := suggestionUC.NewListMemberSkillsUseCase(suggestionService, log)
	setMemberRoleUC := suggestionUC.NewSetMemberRoleUseCase(suggestionService, log)
	setMemberCapacityUC := suggestionUC.NewSetMemberCapacityUseCase(suggestionService, log)
	suggestionHandlerInstance := handler.NewSuggestionHandler(suggestTeammatesUC, getManpowerUC, setMemberSkillsUC, listMemberSkillsUC, setMemberRoleUC, setMemberCapacityUC, log)

	// Suggestion routes
	api.Get("/tasks/:taskId/suggest-teammates", suggestionHandlerInstance.SuggestTeammates)
	api.Get("/projects/:projectId/manpower", suggestionHandlerInstance.GetManpower)
	api.Get("/projects/:projectId/member-skills", suggestionHandlerInstance.ListMemberSkills)
	api.Put("/projects/:projectId/members/:accountId/skills", suggestionHandlerInstance.SetMemberSkills)
	api.Put("/projects/:projectId/members/:accountId/role", suggestionHandlerInstance.SetMemberRole)
	api.Put("/projects/:projectId/members/:accountId/capacity", suggestionHandlerInstance.SetMemberCapacity)

	// Upload setup
	SetupUploadHandlers(api, db, profileRepository, log)
}

// SetupUploadHandlers sets up upload-related routes and dependencies
func SetupUploadHandlers(api fiber.Router, db *gorm.DB, profileRepo profiles.ProfileRepository, log logger.Logger) {
	// Get environment
	endpoint := os.Getenv("S3_ENDPOINT")
	region := os.Getenv("S3_REGION")
	bucketName := os.Getenv("S3_BUCKET")
	accessKeyID := os.Getenv("S3_ACCESS_KEY")
	secretAccessKey := os.Getenv("S3_SECRET_KEY")
	publicURL := os.Getenv("SUPABASE_PUBLIC_URL")

	// Validate required configuration
	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" || publicURL == "" {
		log.Warn("Supabase Storage configuration incomplete, file upload endpoints will not be available", map[string]interface{}{
			"endpoint":   endpoint != "",
			"access_key": accessKeyID != "",
			"secret_key": secretAccessKey != "",
			"public_url": publicURL != "",
		})
		return
	}

	// Create storage repository
	storageRepo, err := repo.NewFileRepository(endpoint, region, bucketName, accessKeyID, secretAccessKey, publicURL)
	if err != nil {
		log.Error("Failed to initialize storage repository, upload endpoints will not be available", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Create file service
	fileService := file.NewFileService(storageRepo, profileRepo)

	// Create file handler
	fileHandler := handler.NewFileHandler(fileService, log)

	// Register routes
	fileGroup := api.Group("/files/avatar")
	fileGroup.Post("/presign", fileHandler.GeneratePresignedURL)
	fileGroup.Post("/upload", fileHandler.UploadFileS3)

	log.Info("File upload endpoints registered successfully", nil)
}
