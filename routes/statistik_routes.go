package routes

import (
	"rt-management/controllers"

	"github.com/gin-gonic/gin"
)

func SetupDashboardRoutes(api *gin.RouterGroup, controller *controllers.DashboardController) {
	api.GET("/dashboard/statistik", controller.DashboardStatistik) // langsung di api, tanpa group "/dashboard"
	api.GET("/keuangan/statistik", controller.GetStatistikKeuangan)
}

