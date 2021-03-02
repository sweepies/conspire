package controllers

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/sweepyoface/conspire/internal/configuration"
	"github.com/sweepyoface/conspire/pkg/s3util"
)

func Test_Upload(t *testing.T) {
	config := configuration.Configure()

	app := fiber.New()
	app.Post("/", Upload(&config, s3util.New(&config)))

	body := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(body)

	formFileWriter, errCreate := multipartWriter.CreateFormFile("file", "gopher2.png")
	assert.Nil(t, errCreate, "CreateFormFile")

	file, errOpen := os.Open(path.Join("../../", "static", "test", "gopher.png"))
	assert.Nil(t, errOpen, "Open file")

	fileBody, errRead := ioutil.ReadAll(file)
	assert.Nil(t, errRead, "Read file")

	formFileWriter.Write(fileBody)

	multipartWriter.Close()
	file.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	resp, errTest := app.Test(req, int(10*time.Second))

	assert.Nil(t, errTest, "POST")
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode, "Status code")

}
