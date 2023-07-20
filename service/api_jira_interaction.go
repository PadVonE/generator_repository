package service

import (
	"fmt"
	"generator/entity"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (s *Service) ListTask(ctx *gin.Context) {
	// Замените PROJECT_KEY на ключ вашего проекта

	organizationId := ctx.Query("organization_id")
	org := entity.Organization{}

	query := s.DB.Model(&org)

	err := query.Where("id = ?", organizationId).Take(&org).Error
	if err != nil {
		log.Error(err)
	}

	jql := fmt.Sprintf("project = %s", org.JiraProject)

	status := ctx.Query("status")
	if len(status) > 0 {
		jql = jql + fmt.Sprintf(" AND status = '%s'", ctx.Query("status"))
	}

	issues, _, err := s.JiraClient.Issue.Search(jql, nil)

	if err != nil {
		log.Error(err)
	}

	var tasksInfo []entity.TaskInfo

	for _, issue := range issues {
		if issue.Fields.Assignee == nil || issue.Fields.Status == nil {
			continue
		}

		taskInfo := entity.TaskInfo{
			IssueKey:    issue.Key,
			Summary:     issue.Fields.Summary,
			AssigneeImg: issue.Fields.Assignee.AvatarUrls.Four8X48,
			Status:      issue.Fields.Status.Name,
		}

		tasksInfo = append(tasksInfo, taskInfo)
	}

	ctx.JSON(200, gin.H{
		"tasks": tasksInfo,
		"err":   err,
	})

}

func (s *Service) ListStatus(ctx *gin.Context) {
	statuses, _, err := s.JiraClient.Status.GetAllStatuses()
	if err != nil {
		panic(err)
	}
	ctx.JSON(200, gin.H{
		"statuses": statuses,
		"err":      err,
	})
}
