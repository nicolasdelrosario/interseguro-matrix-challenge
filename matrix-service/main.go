package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

//go:embed openapi.yaml
var openapiSpec []byte

//go:embed docs.html
var docsPage []byte

type matrixRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

type statisticsRequest struct {
	Matrices [][][]float64 `json:"matrices"`
}

type statistics struct {
	Min               float64 `json:"min"`
	Max               float64 `json:"max"`
	Sum               float64 `json:"sum"`
	Average           float64 `json:"average"`
	HasDiagonalMatrix bool    `json:"hasDiagonalMatrix"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func newApp(statisticsURL string, client *http.Client) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 1 << 20,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			if fiberError, ok := err.(*fiber.Error); ok {
				status = fiberError.Code
			}
			return c.Status(status).JSON(errorResponse{Error: err.Error()})
		},
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type",
	}))
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/yaml")
		return c.Send(openapiSpec)
	})
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Type("html").Send(docsPage)
	})
	matrices := app.Group("/api/v1/matrices")
	matrices.Post("/rotate", func(c *fiber.Ctx) error {
		matrix, err := parseMatrix(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: err.Error()})
		}
		rotated := rotateClockwise(matrix)
		stats, err := fetchStatistics(client, statisticsURL, [][][]float64{rotated})
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(errorResponse{Error: "statistics service unavailable"})
		}
		return c.JSON(fiber.Map{"rotated": rotated, "statistics": stats})
	})
	matrices.Post("/qr", func(c *fiber.Ctx) error {
		matrix, err := parseMatrix(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: err.Error()})
		}
		q, r, err := factorQR(matrix)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: err.Error()})
		}
		stats, err := fetchStatistics(client, statisticsURL, [][][]float64{q, r})
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(errorResponse{Error: "statistics service unavailable"})
		}
		return c.JSON(fiber.Map{"q": q, "r": r, "statistics": stats})
	})
	return app
}

func parseMatrix(c *fiber.Ctx) ([][]float64, error) {
	var request matrixRequest
	if !strings.HasPrefix(c.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		return nil, fmt.Errorf("content type must be application/json")
	}
	if err := json.Unmarshal(c.Body(), &request); err != nil {
		return nil, fmt.Errorf("invalid JSON body")
	}
	if err := validateMatrix(request.Matrix); err != nil {
		return nil, err
	}
	return request.Matrix, nil
}

func fetchStatistics(client *http.Client, url string, matrices [][][]float64) (statistics, error) {
	var result statistics
	body, err := json.Marshal(statisticsRequest{Matrices: matrices})
	if err != nil {
		return result, err
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return result, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return result, fmt.Errorf("statistics service returned %d", response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

func main() {
	statisticsURL := os.Getenv("STATISTICS_URL")
	if statisticsURL == "" {
		statisticsURL = "http://127.0.0.1:3000/api/v1/statistics"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	app := newApp(statisticsURL, &http.Client{Timeout: 30 * time.Second})
	if err := app.Listen(":" + port); err != nil {
		panic(err)
	}
}
