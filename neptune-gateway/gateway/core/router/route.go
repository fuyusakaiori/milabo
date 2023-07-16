package router

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/fuyusakaiori/gateway/core/controller"
	"github.com/fuyusakaiori/gateway/core/docs"
	"github.com/fuyusakaiori/gateway/core/middleware"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)


func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	// 1. 初始化 swagger
	docs.SwaggerInfo.Title = lib.GetStringConf("base.swagger.title")
	docs.SwaggerInfo.Description = lib.GetStringConf("base.swagger.desc")
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = lib.GetStringConf("base.swagger.host")
	docs.SwaggerInfo.BasePath = lib.GetStringConf("base.swagger.base_path")
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	// 2. 初始化 redis
	redis, err := sessions.NewRedisStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	if err != nil {
		log.Fatalf("redis init err: %v", err)
	}
	// 3. 创建路由
	router := gin.Default()
	// 4. 传入默认使用的中间件
	router.Use(middlewares...)
	// 4.1. ping
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// 4.2. swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 4.3. admin
	adminRouter := router.Group("/admin_controller")
	adminRouter.Use(
		sessions.Sessions("admin", redis),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.TranslationMiddleware(),
	)
	{
		// 路由注册
		controller.AdminRegister(adminRouter)
	}

	return router
}
