package routes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/joelinman-nxp/defrosted/app/data"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func NewMiddleware() fiber.Handler {
	return AuthMiddleware
}

func AuthMiddleware(c *fiber.Ctx) error {
	// check if there is an active session
	session, err := store.Get(c)

	//check if the path is home, if so, allow the request
	if strings.Split(c.Path(), "/")[1] == "auth" {
		return c.Next()
	}

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Not Authorized",
		})
	}
	
	if session.Get(AUTH_KEY) == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Not Authorized",
		})
	}
	return c.Next()
}

func Register(c *fiber.Ctx) error {
	c.Accepts("application/json")
	var u user
	err := c.BodyParser(&u)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error parsing the request",
			"error": err.Error(),
		})
	}
	password, bcErr := bcrypt.GenerateFromPassword([]byte(u.Password), 14)

	if bcErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong:" + bcErr.Error(),
		})
	}
	user := data.User {
		Email: u.Email,
		Password: string(password),
		EmailVerified: false,
		Role: "player",
	}
	err = data.Create(&user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong" + err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
	})
}

func Login(c *fiber.Ctx) error {
	var userData user

	err := c.BodyParser(&userData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing the request",
		})
	}
	var user data.User
	// check if user exists
	if !data.CheckEmail(userData.Email, &user) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Player does not exist with that email, Register now!",
		})
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userData.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid Password provided, Please try again",
		})
	}
	sess, sessErr := store.Get(c)
	if sessErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong" + sessErr.Error(),
		})
	}
	player, err := user.GetPlayer(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong getting player" + err.Error(),
		})
	}
	playerJson, err := json.Marshal(&player)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong getting Player JSON data" + err.Error(),
		})
	}
	sess.Set(AUTH_KEY, true)
	sess.Set(USER_ID, user.ID)
	sess.Set(PLAYER, string(playerJson))
	sessionPlayer := sess.Get(PLAYER)
	json.Unmarshal([]byte(sessionPlayer.(string)), &sessionPlayer)
	fmt.Println(sessionPlayer)
	sessErr = sess.Save()
	if sessErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong" + sessErr.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged in successfully",
		"player": player,
	})
}