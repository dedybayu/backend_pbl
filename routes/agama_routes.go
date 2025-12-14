package routes

import (
	"rt-management/controllers"
	"github.com/gin-gonic/gin"
)

func SetupAgamaRoutes(api *gin.RouterGroup, agamaController *controllers.AgamaController) {
	agama := api.Group("/agama")
	{
		agama.GET("/", agamaController.GetAll)
	}
}
