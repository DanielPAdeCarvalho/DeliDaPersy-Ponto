package handlers

import (
	"deli-ponto/configuration"
	"deli-ponto/database/query"
	"deli-ponto/model"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

func GetPunches(c *gin.Context, dynamoClient *dynamodb.Client, logs configuration.GoAppTools) {
	punches := make([]model.Punch, 0)

	ponto := query.SelectPunch("Bia", *dynamoClient, logs)
	punches = append(punches, ponto)
	ponto = query.SelectPunch("Danilo", *dynamoClient, logs)
	punches = append(punches, ponto)
	ponto = query.SelectPunch("paty", *dynamoClient, logs)
	punches = append(punches, ponto)
	c.IndentedJSON(http.StatusOK, punches)

}

func GetReport(c *gin.Context, dynamoClient *dynamodb.Client, logs configuration.GoAppTools, nome string, mes string) {
	if nome == "Bianca" {
		nome = "Bia"
	} else if nome == "Patricia" {
		nome = "paty"
	}
	ano := time.Now().Year()
	periodo := strconv.Itoa(ano) + "-" + mes
	report := query.SelectReport(nome, periodo, *dynamoClient, logs)
	c.IndentedJSON(http.StatusOK, report)
}

func ResponseOK(c *gin.Context, app configuration.GoAppTools) {
	c.IndentedJSON(http.StatusOK, "Servidor up")
}

func PostPunch(nome string, c *gin.Context, dynamoClient *dynamodb.Client, logs configuration.GoAppTools) {
	query.InsertPunch(dynamoClient, nome, logs)
	response := ("ponto do colaborador " + nome + " batido")
	c.IndentedJSON(http.StatusCreated, response)
}
