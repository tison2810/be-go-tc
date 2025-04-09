package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tison2810/be-go-tc/database"
	"github.com/tison2810/be-go-tc/models"
	"github.com/tison2810/be-go-tc/services"
	"github.com/tison2810/be-go-tc/utils"
)

type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token"`
}

func GoogleAuthHandler(c *fiber.Ctx) error {
	var request struct {
		Code string `json:"code"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")

	tokenURL := "https://oauth2.googleapis.com/token"
	data := fmt.Sprintf(
		"code=%s&client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=authorization_code",
		request.Code, clientID, clientSecret, redirectURI,
	)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get token"})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenRes GoogleTokenResponse
	json.Unmarshal(body, &tokenRes)

	if tokenRes.AccessToken == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve access token"})
	}

	user, err := getGoogleUserInfo(tokenRes.AccessToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user info"})
	}

	dbUser, err := services.FindUserByEmail(database.DB.Db, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	if dbUser == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User not found"})
	}
	if dbUser.FirstName == "" && dbUser.LastName == "" {
		dbUser.FirstName, dbUser.LastName = splitName(user.Name)
		database.DB.Db.Model(&dbUser).Updates(models.User{FirstName: dbUser.FirstName, LastName: dbUser.LastName, Role: "student"})
	}

	jwtToken, _ := utils.GenerateJWT(dbUser.Mail, dbUser.Role)

	return c.JSON(fiber.Map{"token": jwtToken, "user": dbUser})
}

func getGoogleUserInfo(accessToken string) (*models.GoogleUser, error) {
	url := "https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var user models.GoogleUser
	json.Unmarshal(body, &user)

	if user.Email == "" {
		return nil, fmt.Errorf("failed to retrieve user info")
	}

	return &user, nil
}

func splitName(fullName string) (string, string) {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "", ""
	}
	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")
	return firstName, lastName
}
