package src

import (
	"canvas_service/types"
	"canvas_service/utils"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AddCanvas(dbpool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req types.SaveCanvasRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido: " + err.Error()})
			return
		}

		fmt.Printf("Petición recibida para guardar canvas: %+v\n", req)

		if req.CanvasID == nil || *req.CanvasID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El campo 'canvasId' es requerido para guardar."})
			return
		}

		tx, err := dbpool.Begin(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo iniciar la transacción."})
			return
		}
		defer tx.Rollback(context.Background())

		canvasID := *req.CanvasID

		var exists bool
		err = tx.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM canvas WHERE id = $1)", canvasID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "El canvas con el ID proporcionado no existe."})
			return
		}

		var canvasUUID pgtype.UUID
		if err := canvasUUID.Scan(canvasID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ID de canvas inválido."})
			return
		}

		if _, err := tx.Exec(context.Background(), "DELETE FROM shapes WHERE canvas_id = $1", canvasUUID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron eliminar las figuras antiguas."})
			return
		}
		if _, err := tx.Exec(context.Background(), "DELETE FROM layers WHERE canvas_id = $1", canvasUUID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron eliminar las capas antiguas."})
			return
		}

		layerRows := make([][]interface{}, len(req.Layers))
		for i, l := range req.Layers {
			layerRows[i] = []interface{}{l.ID, canvasUUID, l.Name, l.Visible, l.Locked, i} // i es el orden
		}
		_, err = tx.CopyFrom(context.Background(), pgx.Identifier{"layers"}, []string{"id", "canvas_id", "name", "visible", "locked", "layer_order"}, pgx.CopyFromRows(layerRows))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron guardar las capas."})
			return
		}

		rows := make([][]interface{}, len(req.Shapes))
		for i, s := range req.Shapes {
			attrs := map[string]interface{}{
				"x1":     s.X1,
				"y1":     s.Y1,
				"x2":     s.X2,
				"y2":     s.Y2,
				"radius": s.Radius,
			}
			if len(s.Points) > 0 {
				attrs["points"] = s.Points
			}
			attrsJSON, _ := json.Marshal(attrs)

			rows[i] = []interface{}{canvasUUID, req.UserID, s.LayerNumber, s.Type, s.Color, s.StrokeWidth, attrsJSON}
		}

		_, err = tx.CopyFrom(context.Background(),
			pgx.Identifier{"shapes"}, //
			[]string{"canvas_id", "user_id", "layer_number", "type", "color", "stroke_width", "attributes"},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron guardar las figuras."})
			return
		}

		if err := tx.Commit(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo confirmar la transacción."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Canvas guardado exitosamente.",
		})
	}
}

func GetCanvasSVG(dbpool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		canvasID := c.Query("id")
		if canvasID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El parámetro 'id' del canvas es requerido."})
			return
		}

		width, _ := strconv.Atoi(c.DefaultQuery("width", "1000"))
		height, _ := strconv.Atoi(c.DefaultQuery("height", "1000"))

		rows, err := dbpool.Query(context.Background(),
			`SELECT type, color, stroke_width, attributes FROM shapes WHERE canvas_id = $1`,
			canvasID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar las figuras."})
			return
		}
		defer rows.Close()

		var shapesToDraw []types.Shape
		for rows.Next() {
			var s types.Shape
			var attrsJSON []byte
			if err := rows.Scan(&s.Type, &s.Color, &s.StrokeWidth, &attrsJSON); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar una figura."})
				return
			}

			if err := json.Unmarshal(attrsJSON, &s); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al decodificar los atributos de una figura."})
				return
			}

			shapesToDraw = append(shapesToDraw, s)
		}

		if rows.Err() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al iterar sobre las figuras."})
			return
		}

		if len(shapesToDraw) == 0 {

		}

		svgData := utils.GenerateSVGFromShapes(shapesToDraw, width, height)
		c.Data(http.StatusOK, "image/svg+xml; charset=utf-8", []byte(svgData))
	}
}

func GetCanvas(dbpool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		canvasID := c.Query("id")
		if canvasID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El parámetro 'id' del canvas es requerido."})
			return
		}

		// Obtener capas
		layerRows, err := dbpool.Query(context.Background(), `SELECT id, name, visible, locked FROM layers WHERE canvas_id = $1 ORDER BY layer_order ASC`, canvasID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar las capas."})
			return
		}
		defer layerRows.Close()

		layers := make([]types.Layer, 0)
		for layerRows.Next() {
			var l types.Layer
			if err := layerRows.Scan(&l.ID, &l.Name, &l.Visible, &l.Locked); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar una capa."})
				return
			}
			layers = append(layers, l)
		}

		rows, err := dbpool.Query(context.Background(),
			`SELECT user_id, layer_number, type, color, stroke_width, attributes FROM shapes WHERE canvas_id = $1`,
			canvasID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar las figuras."})
			return
		}
		defer rows.Close()

		shapes := make([]types.Shape, 0)
		for rows.Next() {
			var s types.Shape
			var attrsJSON []byte
			var userID string
			if err := rows.Scan(&userID, &s.LayerNumber, &s.Type, &s.Color, &s.StrokeWidth, &attrsJSON); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar una figura."})
				return
			}

			if err := json.Unmarshal(attrsJSON, &s); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al decodificar los atributos de una figura."})
				return
			}
			shapes = append(shapes, s)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al iterar sobre las figuras."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"layers": layers,
			"shapes": shapes,
		})
	}
}

func Deletecanvas(dbpool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		canvasID := c.Query("id")
		if canvasID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El parámetro 'id' del canvas es requerido."})
			return
		}
		_, err := dbpool.Exec(context.Background(), "DELETE FROM canvas WHERE id = $1", canvasID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar el canvas."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Canvas eliminado exitosamente."})
	}
}

func Checksum(dbpool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		canvasID := c.Query("id")
		if canvasID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El parámetro 'id' del canvas es requerido."})
			return
		}

		rows, err := dbpool.Query(context.Background(),
			`SELECT type, color, stroke_width, attributes FROM shapes WHERE canvas_id = $1 ORDER BY id`,
			canvasID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar las figuras para el checksum."})
			return
		}
		defer rows.Close()

		hasher := sha256.New()
		for rows.Next() {
			var s types.Shape
			var attrsJSON []byte
			if err := rows.Scan(&s.Type, &s.Color, &s.StrokeWidth, &attrsJSON); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar una figura para el checksum."})
				return
			}

			shapeData, _ := json.Marshal(s)
			hasher.Write(shapeData)
			hasher.Write(attrsJSON)
		}

		c.JSON(http.StatusOK, gin.H{"checksum": hex.EncodeToString(hasher.Sum(nil))})
	}
}
