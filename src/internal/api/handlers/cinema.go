package handlers

import (
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CinemaHandler struct {
	service *service.CinemaService
	factory *impl.RepositoryFactory
	cache   cache.Cache
}

func NewCinemaHandler(cinemaService *service.CinemaService, factory *impl.RepositoryFactory, cacheLocal cache.Cache) *CinemaHandler {
	return &CinemaHandler{service: cinemaService, factory: factory, cache: cacheLocal}
}

func (h *CinemaHandler) Router(group *gin.RouterGroup) {
	verify := middleware.NewAuthMiddleware(h.cache, h.factory.ApiKey(), h.factory.User(), h.factory.GroupPower(), h.factory.Power())
	cinema := group.Group("/cinema")
	cinema.Use(verify.Verify(), middleware.PowerVerify("file:preview"))
	cinema.GET("/:root_id/home", h.Home)
	cinema.GET("/:root_id/latest", h.Latest)
	cinema.GET("/:root_id/folders/:folder_id/videos", h.FolderVideos)
	cinema.GET("/:root_id/videos/:uf_id", h.VideoDetail)
	cinema.GET("/:root_id/videos/:uf_id/related", h.Related)
	cinema.GET("/:root_id/videos/random", h.RandomVideo)
}

func positivePathInt(c *gin.Context, name string) (int, bool) {
	value, err := strconv.Atoi(c.Param(name))
	return value, err == nil && value > 0
}

func queryInt(c *gin.Context, name string) int {
	value, _ := strconv.Atoi(c.Query(name))
	return value
}

func (h *CinemaHandler) Home(c *gin.Context) {
	rootID, ok := positivePathInt(c, "root_id")
	if !ok {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, "影视文件夹ID无效", nil))
		return
	}
	result, err := h.service.Home(c.Request.Context(), c.GetString("userID"), rootID, queryInt(c, "page"), queryInt(c, "page_size"))
	if err != nil {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.NewJsonResponse(200, "查询成功", result))
}

func (h *CinemaHandler) Latest(c *gin.Context) {
	rootID, ok := positivePathInt(c, "root_id")
	if !ok {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, "影视文件夹ID无效", nil))
		return
	}
	result, err := h.service.Latest(c.Request.Context(), c.GetString("userID"), rootID, queryInt(c, "page"), queryInt(c, "page_size"))
	if err != nil {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.NewJsonResponse(200, "查询成功", result))
}

func (h *CinemaHandler) FolderVideos(c *gin.Context) {
	rootID, rootOK := positivePathInt(c, "root_id")
	folderID, folderOK := positivePathInt(c, "folder_id")
	if !rootOK || !folderOK {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, "文件夹ID无效", nil))
		return
	}
	result, err := h.service.FolderVideos(c.Request.Context(), c.GetString("userID"), rootID, folderID, queryInt(c, "page"), queryInt(c, "page_size"))
	if err != nil {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.NewJsonResponse(200, "查询成功", result))
}

func (h *CinemaHandler) VideoDetail(c *gin.Context) {
	rootID, ok := positivePathInt(c, "root_id")
	if !ok {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, "影视文件夹ID无效", nil))
		return
	}
	result, err := h.service.VideoDetail(c.Request.Context(), c.GetString("userID"), rootID, c.Param("uf_id"))
	if err != nil {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.NewJsonResponse(200, "查询成功", result))
}

func (h *CinemaHandler) Related(c *gin.Context) {
	rootID, ok := positivePathInt(c, "root_id")
	if !ok {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, "影视文件夹ID无效", nil))
		return
	}
	result, err := h.service.Related(c.Request.Context(), c.GetString("userID"), rootID, c.Param("uf_id"), queryInt(c, "page"), queryInt(c, "page_size"))
	if err != nil {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.NewJsonResponse(200, "查询成功", result))
}

func (h *CinemaHandler) RandomVideo(c *gin.Context) {
	rootID, ok := positivePathInt(c, "root_id")
	if !ok {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, "影视文件夹ID无效", nil))
		return
	}
	result, err := h.service.RandomVideo(c.Request.Context(), c.GetString("userID"), rootID)
	if err != nil {
		c.JSON(http.StatusOK, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.NewJsonResponse(200, "查询成功", result))
}
