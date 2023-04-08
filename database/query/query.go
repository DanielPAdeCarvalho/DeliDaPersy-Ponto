package query

import (
	"context"
	"deli-ponto/configuration"
	"deli-ponto/model"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/beevik/ntp"
)

func InsertPunch(dynamoClient *dynamodb.Client, nome string, logs configuration.GoAppTools) {

	//o codigo esta indo no observatorio nacional pegar a data e hora
	datatemp, err := ntp.Time("gps.ntp.br")
	configuration.Check(err, logs)

	//Ajusta a hora para o horario de Fortaleza
	loc, err := time.LoadLocation("America/Fortaleza")
	configuration.Check(err, logs)
	data := datatemp.In(loc).Format("2006-01-02_15:04:05")
	//cria o objeto para ser inserido no banco
	ponto := model.Punch{
		Nome: nome,
		Data: data,
	}

	item, err := attributevalue.MarshalMap(ponto)
	configuration.Check(err, logs)

	_, err = dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("PontoColaborador"),
		Item:      item,
	})
	configuration.Check(err, logs)
}

func SelectPunch(nome string, dynamoClient dynamodb.Client, app configuration.GoAppTools) model.Punch {
	// Build the query input
	filter := expression.Key("Nome").Equal(expression.Value(nome))
	proj := expression.NamesList(expression.Name("Nome"), expression.Name("Data"))
	expr, err := expression.NewBuilder().WithKeyCondition(filter).WithProjection(proj).Build()
	configuration.Check(err, app)

	result, err := dynamoClient.Query(context.Background(), &dynamodb.QueryInput{
		TableName:                 aws.String("PontoColaborador"),
		FilterExpression:          expr.Filter(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
		ScanIndexForward:          aws.Bool(true),
	})
	configuration.Check(err, app)
	//
	var punch model.Punch
	for _, i := range result.Items {
		item := model.Punch{}

		err = attributevalue.UnmarshalMap(i, &item)

		if err != nil {
			log.Fatalf("Got error unmarshalling: %s", err)

		}
		punch = item
	}

	return punch
}
