package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var expMap = NewExpressionsMap()

// Структура для чтения выражения из запроса
type ExpressionBody struct {
	Expression string `json:"expression"`
}

type TaskResultBody struct {
	Id     string  `json:"id"`
	Result float64 `json:"result"`
}

func createExpression(c *gin.Context) {
	var expressionBody ExpressionBody
	// Привязываем тело запроса JSON к структуре Task
	if err := c.BindJSON(&expressionBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Проверим выражение на валидность
	_, valid := createStack(expressionBody.Expression)
	if !valid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid expression"})
		return
	}

	var result, exp = expMap.AddExpression(expressionBody.Expression)

	if result != http.StatusCreated {
		c.JSON(result, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(result, gin.H{"id": exp.Id.String()})
}

func getExpressions(c *gin.Context) {
	response := expMap.GetExpressions()
	c.JSON(http.StatusOK, gin.H{"expressions": response})
}

func getExpression(c *gin.Context) {
	// Извлекаем параметр id из пути
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "id is no set"})
		return
	}

	expression, status := expMap.GetExpression(id)

	c.JSON(status, gin.H{"expression": expression})
}

func getTask(c *gin.Context) {
	task := expMap.getTask()
	if task == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

func setTask(c *gin.Context) {
	var taskBody TaskResultBody
	// Привязываем тело запроса JSON к структуре Task
	if err := c.BindJSON(&taskBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := expMap.setTaskResult(taskBody)
	c.Status(result)
}

func main() {
	router := gin.Default()
	// Чтобы создать задачу, определяем маршрут
	router.POST("/api/v1/calculate", createExpression)
	router.GET("/api/v1/expressions", getExpressions)
	router.GET("/api/v1/expressions/:id", getExpression)
	router.GET("/api/internal/task", getTask)
	router.POST("/api/internal/task", setTask)

	router.Run(":8000")
}
