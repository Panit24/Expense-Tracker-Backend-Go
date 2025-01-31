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
	expense.Date = time.Now()
	db.Create(&expense)
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


