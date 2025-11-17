package user

import (
	"clean-arch/internal/dto"
	"clean-arch/internal/factory"
	"clean-arch/pkg/tracer"
	"clean-arch/pkg/util"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type handler struct {
	service Service
}

func NewHandler(f *factory.Factory) *handler {
	return &handler{
		service: NewService(f),
	}
}

func (h *handler) FindAll(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")
	search := c.DefaultQuery("search", "")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		response := util.APIResponse(fmt.Sprintf("Internal server error %s", err.Error()), http.StatusInternalServerError, "error", nil)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		response := util.APIResponse(fmt.Sprintf("Internal server error %s", err.Error()), http.StatusInternalServerError, "error", nil)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	payload := dto.PayloadBasicTable{
		Limit:  limitInt,
		Offset: offsetInt,
		Search: search,
	}

	res, err := h.service.FindAll(c, payload)
	if err != nil {
		response := util.APIResponse("Failed to get user list", http.StatusBadRequest, "error", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := util.APIResponse("Successfully get user list", http.StatusOK, "success", res)
	tracer.Log(c, "info", "Get User List")
	c.JSON(http.StatusOK, response)
}

func (h *handler) Store(c *gin.Context) {
	var req dto.PayloadUser
	if err := c.ShouldBind(&req); err != nil {
		response := util.APIResponse("Invalid request", http.StatusBadRequest, "error", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := validation.ValidateStruct(&req); err != nil {
		response := util.APIResponse("Validation failed", http.StatusBadRequest, "error", err)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var uploadedFile string

	if req.File != nil {
		fileExt := filepath.Ext(req.File.Filename)
		allowedExt := []string{".jpg", ".png", ".jpeg", ".bmp"}
		if !util.InArrayStr(allowedExt, fileExt) {
			response := util.APIResponse(fmt.Sprintf("File ext not valid, allowed ext is %v", allowedExt), http.StatusUnprocessableEntity, "failed", nil)
			c.JSON(http.StatusUnprocessableEntity, response)
			return
		}

		if req.File.Size > (5 * 1024 * 1024) {
			response := util.APIResponse("File size to large, max size allowed is 5Mb", http.StatusUnprocessableEntity, "failed", nil)
			c.JSON(http.StatusUnprocessableEntity, response)
			return
		}

		fileURL, err := util.SaveFile(req.File)
		if err != nil {
			response := util.APIResponse("Failed to upload file", http.StatusInternalServerError, "error", err.Error())
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		appUrl := util.GetEnv("APP_URL", "http://localhost")
		appPort := util.GetEnv("APP_PORT", "8080")

		baseURL := fmt.Sprintf("%s:%s/", appUrl, appPort)
		sanitizedLink := strings.Replace(fileURL, baseURL, "", 1)

		uploadedFile = sanitizedLink
		req.URL = fileURL
	}

	if err := h.service.Store(c, req); err != nil {
		if req.File != nil {
			_ = util.DeleteFile(uploadedFile)
		}
		response := util.APIResponse("Failed to store user", http.StatusInternalServerError, "error", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := util.APIResponse("Successfully store user", http.StatusOK, "success", nil)
	tracer.Log(c, "info", "Store User")
	c.JSON(http.StatusOK, response)
}

func (h *handler) Update(c *gin.Context) {
	id := c.Param("id")
	intId, _ := strconv.Atoi(id)

	var req dto.PayloadUpdateUser
	if err := c.ShouldBind(&req); err != nil {
		response := util.APIResponse("Invalid request", http.StatusBadRequest, "error", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := validation.ValidateStruct(&req); err != nil {
		response := util.APIResponse("Validation failed", http.StatusBadRequest, "error", err)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var uploadedFile string

	if req.File != nil {
		fileExt := filepath.Ext(req.File.Filename)
		allowedExt := []string{".jpg", ".png", ".jpeg", ".bmp"}
		if !util.InArrayStr(allowedExt, fileExt) {
			response := util.APIResponse(fmt.Sprintf("File ext not valid, allowed ext is %v", allowedExt), http.StatusUnprocessableEntity, "failed", nil)
			c.JSON(http.StatusUnprocessableEntity, response)
			return
		}

		if req.File.Size > (5 * 1024 * 1024) {
			response := util.APIResponse("File size to large, max size allowed is 5Mb", http.StatusUnprocessableEntity, "failed", nil)
			c.JSON(http.StatusUnprocessableEntity, response)
			return
		}

		fileURL, err := util.SaveFile(req.File)
		if err != nil {
			response := util.APIResponse("Failed to upload file", http.StatusInternalServerError, "error", err.Error())
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		appUrl := util.GetEnv("APP_URL", "http://localhost")
		appPort := util.GetEnv("APP_PORT", "8080")

		baseURL := fmt.Sprintf("%s:%s/", appUrl, appPort)
		sanitizedLink := strings.Replace(fileURL, baseURL, "", 1)

		uploadedFile = sanitizedLink
		req.URL = fileURL
	}

	if err := h.service.Update(c, intId, req); err != nil {
		if req.File != nil {
			_ = util.DeleteFile(uploadedFile)
		}
		response := util.APIResponse("Failed to update user", http.StatusInternalServerError, "error", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := util.APIResponse("Successfully update user", http.StatusOK, "success", nil)
	tracer.Log(c, "info", "Update User")
	c.JSON(http.StatusOK, response)
}

func (h *handler) FindOne(c *gin.Context) {
	id := c.Param("id")
	intId, _ := strconv.Atoi(id)

	res, err := h.service.FindOne(c, intId)
	if err != nil {
		response := util.APIResponse("Failed to get detail user", http.StatusInternalServerError, "error", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := util.APIResponse("Successfully get detail user", http.StatusOK, "success", res)
	tracer.Log(c, "info", "Update User")
	c.JSON(http.StatusOK, response)
}

func (h *handler) Delete(c *gin.Context) {
	id := c.Param("id")
	intId, _ := strconv.Atoi(id)

	if err := h.service.Delete(c, intId); err != nil {
		response := util.APIResponse("Failed to delete user", http.StatusInternalServerError, "error", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := util.APIResponse("Successfully delete user", http.StatusOK, "success", nil)
	tracer.Log(c, "info", "Delete User")
	c.JSON(http.StatusOK, response)
}
