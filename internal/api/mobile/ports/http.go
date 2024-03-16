package ports

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	browser_adapters "github.com/twofas/2fas-server/internal/api/browser_extension/adapters"
	"github.com/twofas/2fas-server/internal/api/mobile/adapters"
	"github.com/twofas/2fas-server/internal/api/mobile/app"
	"github.com/twofas/2fas-server/internal/api/mobile/app/command"
	"github.com/twofas/2fas-server/internal/api/mobile/app/queries"
	"github.com/twofas/2fas-server/internal/api/mobile/domain"
	"github.com/twofas/2fas-server/internal/common/api"
	"github.com/twofas/2fas-server/internal/common/logging"
)

type RoutesHandler struct {
	cqrs                   *app.Cqrs
	validator              *validator.Validate
	mobileDeviceRepository domain.MobileDeviceRepository
}

func NewRoutesHandler(
	cqrs *app.Cqrs,
	validate *validator.Validate,
	repository domain.MobileDeviceRepository,
) *RoutesHandler {
	return &RoutesHandler{
		cqrs:                   cqrs,
		validator:              validate,
		mobileDeviceRepository: repository,
	}
}

func (r *RoutesHandler) UpdateMobileDevice(c *gin.Context) {
	cmd := &command.UpdateMobileDevice{}

	c.ShouldBindUri(cmd)
	c.ShouldBindJSON(cmd)

	err := r.validator.Struct(cmd)

	logging.Info("Start command", cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.UpdateMobileDevice.Handle(cmd)

	if err != nil {
		var deviceNotFoundErr adapters.MobileDeviceCouldNotBeFound

		if errors.As(err, &deviceNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.MobileDeviceQuery{
		Id: cmd.Id,
	}

	presenter, err := r.cqrs.Queries.MobileDeviceQuery.Handle(q)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) RegisterMobileDevice(c *gin.Context) {
	id := uuid.New()

	cmd := &command.RegisterMobileDevice{
		Id: id,
	}

	c.ShouldBindJSON(cmd)

	err := r.validator.Struct(cmd)

	logging.Info("Start command", cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.RegisterMobileDevice.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.MobileDeviceQuery{
		Id: id.String(),
	}

	presenter, err := r.cqrs.Queries.MobileDeviceQuery.Handle(q)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) RemoveAllMobileDevices(c *gin.Context) {
	cmd := &command.RemoveAllMobileDevices{}

	r.cqrs.Commands.RemoveAllMobileDevices.Handle(cmd)

	c.JSON(200, api.NewOk("Mobile devices have been removed."))
}

func (r *RoutesHandler) PairMobileWithExtension(c *gin.Context) {
	cmd := &command.PairMobileWithBrowserExtension{}

	c.BindJSON(&cmd)
	c.BindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.PairMobileWithExtension.Handle(cmd)

	if err != nil {
		var conflictError domain.ExtensionHasAlreadyBeenPairedError

		if errors.As(err, &conflictError) {
			c.JSON(409, api.ConflictError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.PairedBrowserExtensionQuery{ExtensionId: cmd.ExtensionId}

	presenter, err := r.cqrs.Queries.PairedBrowserExtensionQuery.Handle(q)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) RemovePairingWithExtension(c *gin.Context) {
	cmd := &command.RemoveDevicePairedExtension{}

	c.BindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.RemovePairingWithExtension.Handle(cmd)

	if err != nil {
		var deviceNotFoundErr *adapters.MobileDeviceCouldNotBeFound
		var extensionsNotFoundErr *browser_adapters.BrowserExtensionsCouldNotBeFound

		if errors.As(err, &deviceNotFoundErr) || errors.As(err, &extensionsNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, api.NewOk("Extension has been disconnected from device."))
}

func (r *RoutesHandler) FindAllMobileAppExtensions(c *gin.Context) {
	cmd := &query.DeviceBrowserExtensionsQuery{}

	c.BindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	deviceId, _ := uuid.Parse(cmd.DeviceId)
	_, err = r.mobileDeviceRepository.FindById(deviceId)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	result, err := r.cqrs.Queries.DeviceBrowserExtensionsQuery.Handle(cmd)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) FindMobileAppExtensionById(c *gin.Context) {
	cmd := &query.DeviceBrowserExtensionsQuery{}

	c.BindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	deviceId, _ := uuid.Parse(cmd.DeviceId)
	_, err = r.mobileDeviceRepository.FindById(deviceId)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	result, err := r.cqrs.Queries.DeviceBrowserExtensionsQuery.Handle(cmd)

	if len(result) == 0 {
		c.JSON(404, api.NotFoundError(browser_adapters.BrowserExtensionsCouldNotBeFound{ExtensionId: cmd.ExtensionId}))
		return
	}

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, result[0])
}

func (r *RoutesHandler) Send2FaToken(c *gin.Context) {
	cmd := &command.Send2FaToken{}

	c.BindJSON(&cmd)
	c.BindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	deviceId, _ := uuid.Parse(cmd.DeviceId)
	_, err = r.mobileDeviceRepository.FindById(deviceId)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	err = r.cqrs.Commands.Send2FaToken.Handle(c.Request.Context(), cmd)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, api.NewOk("Token has been sent to browser extension"))
}

func (r *RoutesHandler) GetAll2FaTokenRequests(c *gin.Context) {
	q := &query.DeviceBrowserExtension2FaRequestQuery{}
	c.BindUri(&q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	deviceId, _ := uuid.Parse(q.DeviceId)
	_, err = r.mobileDeviceRepository.FindById(deviceId)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	result, err := r.cqrs.Queries.DeviceBrowserExtension2FaRequestQuery.Handle(q)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))

		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) CreateMobileNotification(c *gin.Context) {
	id := uuid.New()

	cmd := &command.CreateNotification{Id: id}

	c.BindJSON(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.CreateNotification.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.MobileNotificationsQuery{Id: id.String()}

	result, err := r.cqrs.Queries.MobileNotificationsQuery.FindOne(q)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) UpdateMobileNotification(c *gin.Context) {
	cmd := &command.UpdateNotification{}

	c.BindUri(cmd)
	c.BindJSON(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.UpdateNotification.Handle(cmd)

	if err != nil {
		var notificationNotFoundErr *adapters.MobileNotificationCouldNotBeFound

		if errors.As(err, &notificationNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.MobileNotificationsQuery{Id: cmd.Id}

	result, err := r.cqrs.Queries.MobileNotificationsQuery.FindOne(q)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) FindAllMobileNotifications(c *gin.Context) {
	q := &query.MobileNotificationsQuery{}

	c.BindUri(&q)
	c.BindQuery(&q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.MobileNotificationsQuery.FindAll(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) FindMobileNotification(c *gin.Context) {
	q := &query.MobileNotificationsQuery{}

	c.BindUri(&q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.MobileNotificationsQuery.FindOne(q)

	if err != nil {
		var notificationNotFoundErr adapters.MobileNotificationCouldNotBeFound

		if errors.As(err, &notificationNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) RemoveMobileNotification(c *gin.Context) {
	cmd := &command.DeleteNotification{}

	c.BindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.DeleteNotification.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.MobileNotificationCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, api.NewOk("Notification has been removed."))
}

func (r *RoutesHandler) RemoveAllMobileNotifications(c *gin.Context) {
	cmd := &command.DeleteAllNotifications{}

	r.cqrs.Commands.RemoveAllMobileNotifications.Handle(cmd)

	c.JSON(200, api.NewOk("Mobile notifications has been removed."))
}

func (r *RoutesHandler) PublishMobileNotification(c *gin.Context) {
	cmd := &command.PublishNotification{}

	c.BindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.PublishNotification.Handle(cmd)

	if err != nil {
		var notificationNotFoundErr adapters.MobileNotificationCouldNotBeFound

		if errors.As(err, &notificationNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.MobileNotificationsQuery{Id: cmd.Id}

	result, err := r.cqrs.Queries.MobileNotificationsQuery.FindOne(q)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	c.JSON(200, result)
}
