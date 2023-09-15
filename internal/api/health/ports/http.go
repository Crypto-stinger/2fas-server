package ports

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/twofas/2fas-server/config"
	mobile "github.com/twofas/2fas-server/internal/api/mobile/domain"
	support "github.com/twofas/2fas-server/internal/api/support/domain"
	"github.com/twofas/2fas-server/internal/common/aws"
	"github.com/twofas/2fas-server/internal/common/clock"
	"github.com/twofas/2fas-server/internal/common/logging"
)

type RoutesHandler struct {
	applicationName string
	redis           *redis.Client
}

func NewRoutesHandler(applicationName string, redis *redis.Client) *RoutesHandler {
	return &RoutesHandler{
		applicationName: applicationName,
		redis:           redis,
	}
}

func (r *RoutesHandler) CheckApplicationHealth(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func (r *RoutesHandler) FakeError(c *gin.Context) {
	logging.WithFields(logging.Fields{
		"context_field_1": "some value",
		"context_field_2": "another value",
	}).Error("Standard fake error")

	messageStruct := &struct {
		Ctx1    string `json:"ctx_1"`
		Ctx2    string `json:"ctx_2"`
		Message string `json:"message"`
	}{
		Ctx1:    "msg json key 1",
		Ctx2:    "msg json key 2",
		Message: "Fake error with message as JSON",
	}

	message, _ := json.Marshal(messageStruct)

	logging.Error(string(message))

	c.JSON(500, gin.H{})
}

func (r *RoutesHandler) FakeWarning(c *gin.Context) {
	logging.WithFields(logging.Fields{
		"context_field_1": "some value",
		"context_field_2": "another value",
	}).Warning("Fake warning")

	c.JSON(200, gin.H{})
}

func (r *RoutesHandler) FakeSecurityWarning(c *gin.Context) {
	logging.WithFields(logging.Fields{
		"type": "security",
		"ip":   c.ClientIP(),
		"uri":  c.Request.URL.String(),
	}).Warning("Fake warning")

	c.JSON(200, gin.H{})
}

type configuration struct {
	Common                 config.Configuration    `json:"common"`
	DebugLogsConfiguration support.DebugLogsConfig `json:"mobile_debug"`
	PushConfiguration      *mobile.FcmPushConfig   `json:"push"`
}

type redisStatus struct {
	Addr       string `json:"addr"`
	Db         int    `json:"db"`
	Connection string `json:"connection"`
}

type systemInfo struct {
	ApplicationName string        `json:"application_name"`
	LocalTime       string        `json:"local_time"`
	Config          configuration `json:"configuration"`
	Environment     []string      `json:"environment"`
	Redis           redisStatus   `json:"redis"`
}

func (r *RoutesHandler) RedisInfo(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	res := r.redis.Info(ctx)

	c.String(200, res.String())
}

func (r *RoutesHandler) GetApplicationConfiguration(c *gin.Context) {
	s3 := aws.NewAwsS3(config.Config.Aws.Region, config.Config.Aws.S3AccessKeyId, config.Config.Aws.S3AccessSecretKey)
	pushConfig := mobile.NewFcmPushConfig(s3)
	debugLogsConfig := support.LoadDebugLogsConfig()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	res := r.redis.Info(ctx)

	var info redisStatus

	if res.Err() != nil {
		info = redisStatus{Connection: res.Err().Error()}
	} else {
		info = redisStatus{Connection: "OK"}
	}

	info.Addr = r.redis.Options().Addr
	info.Db = r.redis.Options().DB

	c.JSON(200, &systemInfo{
		LocalTime: clock.New().Now().String(),
		Config: configuration{
			Common:                 config.Config,
			DebugLogsConfiguration: debugLogsConfig,
			PushConfiguration:      pushConfig,
		},
		Environment: os.Environ(),
		Redis:       info,
	})
}
