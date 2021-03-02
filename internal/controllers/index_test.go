package controllers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func Test_Index(t *testing.T) {
	app := fiber.New()

	app.Get("/", Index("../../static"))
	req := httptest.NewRequest("GET", "/", nil)

	resp, err := app.Test(req)
	assert.Nil(t, err, "GET")
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Status code")
	assert.Greater(t, resp.ContentLength, int64(0), "Content length")
}
