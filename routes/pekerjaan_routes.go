package routes

import (
	"rt-management/controllers"
	"github.com/gin-gonic/gin"
)

func SetupPekerjaanRoutes(api *gin.RouterGroup, pekerjaanController *controllers.PekerjaanController) {
	pekerjaan := api.Group("/pekerjaan")
	{
		pekerjaan.GET("/", pekerjaanController.GetAll)
	}
}
