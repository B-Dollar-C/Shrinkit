package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	collection *mongo.Collection
)

type Redirection struct {
	ID          string    `bson:"_id,omitempty" json: "id"`
	OriginalURL string    `bson:"originalurl" json: "originalUrl"`
	NewURL      string    `bson:"newurl" json: "newUrl"`
	CreatedAt   time.Time `bson:"createdat" json: "createdAt"`
	ShortUrl    string    `bson: "shorturl json: "shortUrl"`
}

func initMongoDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(context.Background(), clientOptions)
	collection = client.Database("url_shortener").Collection("redirections")
}

func redirectToNewURL(c *fiber.Ctx) error {
	shortUrl := c.Params("shortUrl")
	fmt.Println(shortUrl)

	var redirection Redirection

	err := collection.FindOne(context.Background(), bson.M{"shorturl": shortUrl}).Decode(&redirection)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("URL not found")
	}

	return c.Redirect(redirection.OriginalURL, http.StatusSeeOther)
}

func generateShortURL(c *fiber.Ctx) error {
	var redirection Redirection

	if err := c.BodyParser(&redirection); err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid request body")
	}

	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	randomString := ""
	for i := 0; i < 5; i++ {
		randomString += string(chars[rand.Intn(len(chars))])
	}

	redirection.ShortUrl = randomString
	redirection.NewURL = "http://localhost:8080/" + randomString

	_, err := collection.InsertOne(context.Background(), redirection)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error saving data")
	}

	return c.Status(http.StatusCreated).JSON(redirection)
}

func main() {
	initMongoDB()

	app := fiber.New()
	app.Static("/", "./index.html")

	corsConfig := cors.New(cors.Config{
		AllowOrigins: "https://shrinkit-ashy.vercel.app", // Allow requests from your domain
	})
	app.Use(corsConfig)

	app.Get("/:shortUrl", redirectToNewURL)
	app.Post("/shorten", generateShortURL)

	fmt.Println("Server started on :8080")
	app.Listen(":8080")
}
