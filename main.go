package main

import (
	"canvas_service/database"
	"canvas_service/messaging"
	"canvas_service/src"
	"fmt"
	"log"

	//"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Conectar a la base de datos
	dbpool, err := database.Connect()
	if err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}
	defer dbpool.Close()

	// Configurar el router de Gin
	router := gin.Default()

	// Configurar CORS
	//config := cors.DefaultConfig()
	//config.AllowAllOrigins = true
	//config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	//config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	//router.Use(cors.New(config))

	// Iniciar el consumidor de RabbitMQ en una goroutine para que no bloquee el servidor.
	go messaging.StartConsumer(dbpool)

	// Definir las rutas
	// La ruta POST /canvas se elimina, ya que la creación ahora es asíncrona.
	router.POST("/canvas/save", src.AddCanvas(dbpool))
	router.GET("/canvas", src.GetCanvas(dbpool))
	router.GET("/canvas/svg", src.GetCanvasSVG(dbpool))
	router.GET("/canvas/checksum", src.Checksum(dbpool))
	router.DELETE("/canvas", src.Deletecanvas(dbpool))

	fmt.Println("Servidor Canvas_Service corriendo en el puerto 8080")
	router.Run(":8080")
}
