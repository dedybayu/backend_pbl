package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupTransaksiRoutes(api *gin.RouterGroup, transaksiController *controllers.TransaksiController, detailTransaksiController *controllers.DetailTransaksiController, authMiddleware *middleware.AuthMiddleware) {
	// Group untuk transaksi
	transaksi := api.Group("/transaksi")
	transaksi.Use(authMiddleware.Auth())
	{
		// Buat transaksi baru dari keranjang
		transaksi.POST("", transaksiController.CreateTransaksi)
		
		// Dapatkan semua transaksi (admin lihat semua, user lihat sendiri)
		transaksi.GET("", transaksiController.GetAllTransaksi)
		
		// Dapatkan transaksi user sendiri
		transaksi.GET("/my", transaksiController.GetTransaksiByUser)
		
		// Dapatkan transaksi by ID
		transaksi.GET("/:id", transaksiController.GetTransaksiByID)
		
		// Update transaksi (hanya admin)
		transaksi.PUT("/:id", authMiddleware.RequireLevel(1), transaksiController.UpdateTransaksi)
		
		// Statistik transaksi (hanya admin)
		transaksi.GET("/stats", authMiddleware.RequireLevel(1), transaksiController.GetTransaksiStats)
	}
	
	// Group untuk detail transaksi
	detail := api.Group("/detail-transaksi")
	detail.Use(authMiddleware.Auth())
	{
		// Dapatkan detail by transaksi ID
		detail.GET("/transaksi/:transaksi_id", detailTransaksiController.GetDetailByTransaksiID)
		
		// Dapatkan detail by ID
		detail.GET("/:id", detailTransaksiController.GetDetailByID)
	}
}