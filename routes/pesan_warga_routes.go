package routes

import (
	"rt-management/controllers"

	"github.com/gin-gonic/gin"
)

func SetupPesanWargaRoutes(api *gin.RouterGroup, pesanController *controllers.PesanWargaController) {
	pesan := api.Group("/pesan-warga")
	{
		pesan.POST("/", pesanController.Create)
		pesan.PUT("/:id", pesanController.Update)
		pesan.GET("/", pesanController.GetAll)
		pesan.GET("/:id", pesanController.GetByID)
		pesan.GET("/user/:user_id", pesanController.GetByUserID)
	}
}
