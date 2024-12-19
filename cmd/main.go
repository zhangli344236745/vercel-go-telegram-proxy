package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	bot *tgbotapi.BotAPI
}

func NewTelegramBot(token string) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	// 设置日志级别
	bot.Debug = true

	return &TelegramBot{bot: bot}, nil
}

// handleUpdate 处理传入的Telegram更新
func (tb *TelegramBot) handleUpdate(update tgbotapi.Update) {
	// 忽略没有消息的更新
	if update.Message == nil {
		return
	}

	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	// 消息回复逻辑
	var reply string
	switch update.Message.Text {
	case "/start":
		reply = "欢迎！我是一个使用Gin框架开发的Telegram机器人。"
	case "/help":
		reply = "可用命令：\n/start - 开始\n/help - 帮助\n/info - 机器人信息"
	case "/info":
		reply = fmt.Sprintf("机器人名称：%s\n用户名：%s", tb.bot.Self.FirstName, tb.bot.Self.UserName)
	default:
		reply = "未知命令。输入 /help 查看可用命令。"
	}

	// 发送回复消息
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	msg.ReplyToMessageID = update.Message.MessageID

	if _, err := tb.bot.Send(msg); err != nil {
		log.Printf("发送消息时发生错误: %v", err)
	}
}

func main() {
	// 从环境变量读取Telegram Bot Token
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("未设置TELEGRAM_BOT_TOKEN环境变量")
	}
	log.Printf("hello,%s", "hello")
	// 创建Telegram机器人
	telegramBot, err := NewTelegramBot(token)
	if err != nil {
		log.Fatalf("创建Telegram机器人失败: %v", err)
	}

	// 创建Gin路由器
	router := gin.Default()

	// Webhook处理器
	router.POST("/webhook", func(c *gin.Context) {
		var update tgbotapi.Update
		if err := c.BindJSON(&update); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 处理更新
		telegramBot.handleUpdate(update)

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"botName":  telegramBot.bot.Self.FirstName,
			"username": telegramBot.bot.Self.UserName,
		})
	})

	// 设置Webhook
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("未设置WEBHOOK_URL环境变量")
	}

	webhookConfig, _ := tgbotapi.NewWebhook(webhookURL + "/webhook")
	_, err = telegramBot.bot.Request(webhookConfig)
	if err != nil {
		log.Fatalf("设置Webhook失败: %v", err)
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("服务器启动，监听端口 %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
