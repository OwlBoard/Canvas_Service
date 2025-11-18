package main

import (
	"canvas_service/database"
	"canvas_service/messaging"
	"canvas_service/src"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbpool *pgxpool.Pool
	ready  uint32 // 0 = not ready, 1 = ready
)

func main() {
	// Configurar el router de Gin
	router := gin.Default()

<<<<<<< Updated upstream
	// Configurar CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))
=======
	// Health endpoint: returns 200 only when the service is ready
	router.GET("/health", func(c *gin.Context) {
		if atomic.LoadUint32(&ready) == 1 {
			c.String(http.StatusOK, "OK")
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "starting"})
	})
>>>>>>> Stashed changes

	// Readiness middleware: block all non-health requests until ready
	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		if atomic.LoadUint32(&ready) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service not ready"})
			c.Abort()
			return
		}
		c.Next()
	})

	// Register handlers that will use the DB pool once available.
	// They are wrapped so they will only execute when `ready == 1`.
	router.POST("/canvas/save", func(c *gin.Context) { src.AddCanvas(dbpool)(c) })
	router.GET("/canvas", func(c *gin.Context) { src.GetCanvas(dbpool)(c) })
	router.GET("/canvas/svg", func(c *gin.Context) { src.GetCanvasSVG(dbpool)(c) })
	router.GET("/canvas/checksum", func(c *gin.Context) { src.Checksum(dbpool)(c) })
	router.DELETE("/canvas", func(c *gin.Context) { src.Deletecanvas(dbpool)(c) })

	// Background goroutine: attempt to connect to DB with retries and backoff.
	// This loop will retry indefinitely (no fatal exit) so the container remains up
	// and /health returns 503 until the DB becomes available.
	go func() {
		wait := 1 * time.Second
		attempt := 0
		for {
			attempt++
			pool, err := database.Connect()
			if err == nil {
				dbpool = pool
				// Start the consumer now that DB is available
				go messaging.StartConsumer(dbpool)
				atomic.StoreUint32(&ready, 1)
				fmt.Printf("Canvas_Service: ready and connected to DB (after %d attempts)\n", attempt)
				return
			}
			log.Printf("Canvas_Service: DB connect attempt %d failed: %v. Retrying in %s", attempt, err, wait)
			time.Sleep(wait)
			wait *= 2
			if wait > 30*time.Second {
				wait = 30 * time.Second
			}
			// continue retrying indefinitely
		}
	}()

	fmt.Println("Servidor Canvas_Service inicializando (health on /health).")
	router.Run(":8080")
}
