package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

type DeliveryRoute string
type NotificationType string

const (
	Email  DeliveryRoute = "email"
	System DeliveryRoute = "system"

	Instant NotificationType = "instant"
	Batch   NotificationType = "batch"
)

type Notification struct {
	Date             string           `json:"date"`
	EventName        string           `json:"eventName"`
	DeliveryRoute    DeliveryRoute    `json:"deliveryRoute"`
	NotificationType NotificationType `json:"notificationType"`
	Metadata         json.RawMessage  `json:"metadata"`
}

// I believe it would be better to use the metadata as just the directory but the assignment said that the body of was a part of the metadata

type EmailMetadata struct {
	EmailAddress string `json:"emailAddress"`
	EmailBody    string `json:"emailBody"`
}

type SystemMetadata struct {
	UUID string `json:"uuid"`
	Body string `json:"body"`
}

func main() {
	r := gin.Default()

	// Imagine this map are several SQS queues by topic, otherwise there isn't persistence
	var notificationsMap = make(map[string][]EmailMetadata)

	// Email service Dependency Injection
	var emailService EmailService
	var notificationRepo NotificationRepository
	var batchAmount int

	switch os.Getenv("EMAIL_PROTOCOL") {
	case "SMTP":
		emailService = SMTPService{}
	case "OTHER":
		emailService = OtherProtocolService{}
	default:
		log.Fatal("Invalid email protocol")
	}

	// Same logic as above for the notificationRepo and batchAmount

	// Endpoints

	r.POST("/notification", func(c *gin.Context) {
		var notification Notification
		if err := c.ShouldBindJSON(&notification); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		switch notification.DeliveryRoute {
		case Email:
			var emailMetadata EmailMetadata
			if err := json.Unmarshal(notification.Metadata, &emailMetadata); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email metadata"})
				return
			}

			switch notification.NotificationType {
			case Instant:
				if err := emailService.SendEmail(emailMetadata.EmailAddress, emailMetadata.EmailBody); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
					return
				}
			case Batch:
				// Add the notification to the queue accordingly
				notificationsMap[notification.EventName] = append(notificationsMap[notification.EventName], emailMetadata)

				// Check if there are batchAmount notifications for this event
				if len(notificationsMap[notification.EventName]) == batchAmount {
					// Join all the email bodies together
					var emailBody string
					for _, email := range notificationsMap[notification.EventName] {
						emailBody += email.EmailBody + "\n"
					}

					// Send the email
					if err := emailService.SendEmail(notificationsMap[notification.EventName][0].EmailAddress, emailBody); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
						return
					}

					// Clear the notifications for this event
					notificationsMap[notification.EventName] = []EmailMetadata{}
				}
			}
		case System:
			// Handle system notification type here
			if err := notificationRepo.Insert(notification); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert notification"})
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification type"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Notification received"})
	})

	r.GET("/notifications", func(c *gin.Context) {
		notifications, err := notificationRepo.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
			return
		}
		c.JSON(http.StatusOK, notifications)
	})

	r.DELETE("/notifications/:id", func(c *gin.Context) {
		id := c.Param("id")
		err := notificationRepo.Remove(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove notification"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Notification removed"})
	})

	r.PUT("/notifications/:id/read", func(c *gin.Context) {
		id := c.Param("id")
		err := notificationRepo.MarkAsRead(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
	})

	r.Run()
}

// Email Service

type EmailService interface {
	SendEmail(address string, body string) error
}

type SMTPService struct{}

func (s SMTPService) SendEmail(address string, body string) error {
	// Implement SMTP email sending here
	return nil
}

type OtherProtocolService struct{}

func (o OtherProtocolService) SendEmail(address string, body string) error {
	// Implement other protocol email sending here
	return nil
}

// Notification Repository

type NotificationRepository interface {
	GetAll() ([]Notification, error)
	Remove(id string) error
	MarkAsRead(id string) error
	Insert(item Notification) error
}

type NotificationRepositoryImpl struct{}

func (n NotificationRepositoryImpl) GetAll() ([]Notification, error) {
	// Implement the method to get all notifications here
	return nil, nil
}

func (n NotificationRepositoryImpl) Remove(id string) error {
	// Implement the method to remove a notification here
	return nil
}

func (n NotificationRepositoryImpl) Insert(item Notification) error {
	// Implement the method to insert a notification here
	return nil
}

func (n NotificationRepositoryImpl) MarkAsRead(id string) error {
	// Implement the method to mark a notification as read here
	return nil
}
