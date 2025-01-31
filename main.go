package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"net/http"


	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

var db *gorm.DB

// Expense model
type Expense struct {
	ID     uint    `gorm:"primaryKey" json:"id"`
	Title   string  `json:"title"`
	Description   string  `json:"description"`
	Category   string  `json:"category"`
	Amount float64 `json:"amount"`
	Date   time.Time  `json:"date"`
}

// Initialize DB
func initDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var errDB error
	db, errDB = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if errDB != nil {
		log.Fatal("Failed to connect to database:", errDB)
	}

	// Auto-migrate the Expense table
	db.AutoMigrate(&Expense{})
	fmt.Println("Database connected and migrated successfully!")
}

func main() {
	initDB()
	r := gin.Default()

	// Configure CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins (change this for security)
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Apply middleware
	r.Use(func(c *gin.Context) {
		corsMiddleware.HandlerFunc(c.Writer, c.Request)
		c.Next()
	})

	// API routes
	r.POST("/expenses", createExpense)
	r.GET("/expenses", getExpenses)
	r.GET("/expenses/:id", getExpenseByID)
	r.PUT("/expenses/:id", updateExpense)
	r.DELETE("/expenses/:id", deleteExpense)

	fmt.Printf("Server running on http://localhost:8000\n")

	// Run the server
	r.Run(":8000")
}

func createExpense(c *gin.Context) {
	var expense Expense
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// If the date is provided and is in the incomplete format (YYYY-MM-DDTHH:MM),
	// add seconds and a timezone offset to complete the date
	if expense.Date.IsZero() && expense.Date.String() != "" {
		// Assuming the date is in the format "2025-01-17T23:51"
		// Add missing seconds and timezone offset (UTC)
		parsedDate, err := time.Parse("2006-01-02T15:04", expense.Date.String())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		// Add seconds (00) and assume UTC timezone
		expense.Date = parsedDate.UTC().Add(time.Second * 0)
	}

	// Store expense in the database (assuming db is a valid database connection)
	if err := db.Create(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense"})
		return
	}
  //---
	// expense.Date = time.Now()
	// db.Create(&expense)
	//---
	c.JSON(http.StatusCreated, expense)
}

// Get All Expenses
func getExpenses(c *gin.Context) {
	var expenses []Expense
	db.Find(&expenses)
	c.JSON(http.StatusOK, expenses)
}

// Get Single Expense
func getExpenseByID(c *gin.Context) {
	id := c.Param("id")
	var expense Expense
	if err := db.First(&expense, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}
	c.JSON(http.StatusOK, expense)
}

// Update Expense
func updateExpense(c *gin.Context) {
	id := c.Param("id")
	var expense Expense
	if err := db.First(&expense, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}

	var updatedExpense Expense
	if err := c.ShouldBindJSON(&updatedExpense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Model(&expense).Updates(updatedExpense)
	c.JSON(http.StatusOK, expense)
}

// Delete Expense
func deleteExpense(c *gin.Context) {
	id := c.Param("id")
	var expense Expense
	if err := db.First(&expense, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}

	db.Delete(&expense)
	c.JSON(http.StatusOK, gin.H{"message": "Expense deleted"})
}


