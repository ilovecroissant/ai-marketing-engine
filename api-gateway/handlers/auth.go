package handlers

import (
    "context"
    "net/http"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/ilovecroissant/ai-marketing-engine/db"
    "golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

func Register(c *gin.Context) {
    var input RegisterInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
        return
    }

    var id string
    err = db.Pool.QueryRow(context.Background(),
        `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id`,
        input.Email, string(hash),
    ).Scan(&id)
    if err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"user_id": id})
}

func Login(c *gin.Context) {
    var input RegisterInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var id, hash string
    err := db.Pool.QueryRow(context.Background(),
        `SELECT id, password FROM users WHERE email = $1`, input.Email,
    ).Scan(&id, &hash)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub": id,
        "exp": time.Now().Add(24 * time.Hour).Unix(),
    })
    signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not sign token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": signed})
}