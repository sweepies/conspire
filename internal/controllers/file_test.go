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
	"github.com/sweepyoface/conspire/pkg/s3util"
)

func Test_File(t *testing.T) {
	configuration.ConfigureTest()

	app := fiber.New()
	app.Get("/:file", File(s3util.New(), true))

	req := httptest.NewRequest("GET", "/gopher.png", nil)

	resp, requestErr := app.Test(req)
	assert.Nil(t, requestErr, "GET")
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Status code")
	assert.Greater(t, resp.ContentLength, int64(0), "Content length")

	body, readErr := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	assert.Nil(t, readErr, "Read body")

	testSum := "83a2a4dfff717653df00fb25984f5c95fcf5ee0934c842b2f7a66b0bf39c4f3e"
	sum := sha256.Sum256(body)

	assert.Equal(t, testSum, hex.EncodeToString(sum[:]), "Checksum")
}
