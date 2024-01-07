package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ueba-profiling/manager"
)

type JobHandler struct {
	JobManager *manager.JobManager
}

func NewJobHandler(jobManager *manager.JobManager) *JobHandler {
	return &JobHandler{
		JobManager: jobManager,
	}
}

func (s *JobHandler) MakeHandler(g *gin.RouterGroup) {
	group := g.Group("/jobs")
	group.GET("", s.getAllJobs)
	group.GET("/succeed", s.getAllJobs)
	group.GET("/failed", s.getAllJobs)
}

func (s *JobHandler) getAllJobs(c *gin.Context) {
	stats := s.JobManager.GetJobInfo()
	c.JSON(http.StatusOK, stats)
}
