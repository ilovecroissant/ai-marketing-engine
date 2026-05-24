package handlers

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/ilovecroissant/ai-marketing-engine/db"
    appkafka "github.com/ilovecroissant/ai-marketing-engine/kafka"
)

type CampaignInput struct {
    Product  string `json:"product"  binding:"required"`
    Industry string `json:"industry" binding:"required"`
    Tone     string `json:"tone"`
    Platform string `json:"platform"`
    Goal     string `json:"goal"`
    Website  string `json:"website"`
}

func CreateCampaign(c *gin.Context) {
    var input CampaignInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID := c.GetString("userID")

    var id string
    err := db.Pool.QueryRow(context.Background(),
        `INSERT INTO campaigns (user_id, product, industry, tone, platform, goal, website)
         VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
        userID, input.Product, input.Industry, input.Tone, input.Platform, input.Goal, input.Website,
    ).Scan(&id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
        return
    }

    // Publish to Kafka (downstream crawler/content-gen will consume this)
    _ = appkafka.PublishCampaign(c.Request.Context(), map[string]any{
        "campaign_id": id,
        "industry":    input.Industry,
        "product":     input.Product,
    })

    c.JSON(http.StatusCreated, gin.H{"campaign_id": id})
}

func GetCampaign(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetString("userID")

    var product, industry, status string
    err := db.Pool.QueryRow(context.Background(),
        `SELECT product, industry, status FROM campaigns WHERE id=$1 AND user_id=$2`,
        id, userID,
    ).Scan(&product, &industry, &status)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"id": id, "product": product, "industry": industry, "status": status})
}