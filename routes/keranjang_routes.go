package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupKeranjangRoutes(api *gin.RouterGroup, keranjangController *controllers.KeranjangController, authMiddleware *middleware.AuthMiddleware) {
	keranjang := api.Group("/keranjang")
	keranjang.Use(authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6)) // Semua endpoint keranjang butuh auth
	{
		// Tambah ke keranjang
		keranjang.POST("", keranjangController.AddToCart)
		
		// Lihat keranjang
		keranjang.GET("", keranjangController.GetCartItems)
		
		// Update item keranjang
		keranjang.PUT("/:id", keranjangController.UpdateCartItem)
		
		// Hapus item dari keranjang
		keranjang.DELETE("/:id", keranjangController.RemoveFromCart)
		
		// Kosongkan keranjang
		keranjang.DELETE("/clear", keranjangController.ClearCart)
		
		// Jumlah item keranjang
		keranjang.GET("/count", keranjangController.GetCartCount)
		
		// Preview checkout
		keranjang.GET("/checkout-preview", keranjangController.CheckoutPreview)
	}
}