package controllers

import (
	"crypto/sha256"
	"io/ioutil"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func Test_Favicon(t *testing.T) {
	app := fiber.New()

	app.Get("/", Favicon("../../static"))

	req := httptest.NewRequest("GET", "/", nil)

	resp, err := app.Test(req)

	assert.Nil(t, err, "GET")
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Status code")

	respBytes, errRespRead := ioutil.ReadAll(resp.Body)
	fileBytes, errFileRead := ioutil.ReadFile(filepath.Join("../../static", "favicon", "default.ico"))

	assert.NoError(t, errRespRead)
	assert.NoError(t, errFileRead)

	respSum := sha256.Sum256(respBytes)
	fileSum := sha256.Sum256(fileBytes)

	assert.Equal(t, fileSum, respSum, "Checksum")

}
