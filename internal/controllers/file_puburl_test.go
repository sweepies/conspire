package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/sweepyoface/conspire/internal/configuration"
)

func Test_FilePubURL(t *testing.T) {
	config := configuration.Configure()

	app := fiber.New()
	app.Get("/:file", FilePubURL(&config, true))

	req := httptest.NewRequest("GET", "/gopher.png", nil)

	resp, requestErr := app.Test(req)
	assert.Nil(t, requestErr, "GET")
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Status code")
	assert.Greater(t, resp.ContentLength, int64(0), "Content length")
	assert.NotEmpty(t, resp.Header.Get("Cache-Control"), "Cache Control")
	assert.Equal(t, "image/png", resp.Header.Get("Content-Type"), "Content Type")

	body, readErr := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	assert.Nil(t, readErr, "Read body")

	testSum := "83a2a4dfff717653df00fb25984f5c95fcf5ee0934c842b2f7a66b0bf39c4f3e"
	sum := sha256.Sum256(body)

	assert.Equal(t, testSum, hex.EncodeToString(sum[:]), "Checksum")
}
