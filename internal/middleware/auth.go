package middleware

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/rs/zerolog/log"
)

type user struct {
	Username string
	Password string
}

// TODO: Find a better authentication method

// Auth returns the authentication middleware
func Auth(c chan fiber.Handler) {
	// read json
	data, err := ioutil.ReadFile(filepath.Join("users.json"))

	if err != nil {
		log.Warn().Err(err).Msg("Error reading users file")
	}

	// deserialize
	var users []user
	json.Unmarshal(data, &users)

	// map the data for the middleware
	userMap := make(map[string]string)
	for _, user := range users {
		userMap[user.Username] = user.Password
		log.Debug().Str("user", user.Username).Msg("Configuring privileged user")
	}

	auth := basicauth.New(basicauth.Config{
		Users: userMap,
	})

	c <- auth
}
