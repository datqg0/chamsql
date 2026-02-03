package handlers

import (
	"judge/database"
	"judge/models"
	"net/http"
	//"strconv"

	"github.com/gin-gonic/gin"
)

func GetProblems(c *gin.Context) {
	rows, _ := database.JudgeDB.Query("SELECT id,name,description FROM problems")
	var list []models.Problem

	for rows.Next() {
		var p models.Problem
		rows.Scan(&p.ID, &p.Name, &p.Description)
		list = append(list, p)
	}
	c.JSON(http.StatusOK, list)
}

func GetProblem(c *gin.Context) {
	id := c.Param("id")
	var p models.Problem
	database.JudgeDB.QueryRow("SELECT id,name,description FROM problems WHERE id=?", id).
		Scan(&p.ID, &p.Name, &p.Description)
	c.JSON(http.StatusOK, p)
}
