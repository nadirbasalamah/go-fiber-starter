package api

import (
	"go-fiber-starter/api/handlers"
	"go-fiber-starter/api/middlewares"
	"go-fiber-starter/auth"
	"go-fiber-starter/pkg/fileupload"
	"go-fiber-starter/users"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/helmet"

	zapMiddleware "github.com/gofiber/contrib/v3/zap"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewApp(repository *gorm.DB, cld *cloudinary.Cloudinary, jwtConfig middlewares.JWTConfig) *fiber.App {
	var (
		app      = fiber.New()
		uploader = fileupload.InitUploader(cld)

		authService = auth.New(repository)
		userService = users.New(repository)

		authHandler  = handlers.NewAuth(authService, jwtConfig)
		usersHandler = handlers.NewUsers(userService, uploader)
	)

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logMiddleware := middlewares.LoggerConfig{Config: zapMiddleware.Config{
		Logger: logger,
	}}

	jwtMiddleware := jwtConfig.Init()

	app.Use(logMiddleware.Init())
	app.Use(helmet.New())

	authRoutes := app.Group("/api/v1/auth")

	authRoutes.Post("/register", middlewares.ValidateBody(&auth.RegisterRequest{}), authHandler.Register)
	authRoutes.Post("/login", middlewares.ValidateBody(&auth.LoginRequest{}), authHandler.Login)

	userRoutes := app.Group("/api/v1", jwtMiddleware, middlewares.VerifyToken())

	userRoutes.Get("/profile", usersHandler.GetProfile)
	userRoutes.Patch("/profile/edit", middlewares.ValidateBody(&users.EditProfileRequest{}), usersHandler.EditProfile)

	return app
}
