package query

import (
	"context"
	"deli-ponto/configuration"
	"deli-ponto/model"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/beevik/ntp"
)

// InsertPunch insere um registro de ponto no banco de dados com o nome do colaborador e a data e hora atual
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

// SelectPunch faz uma query no banco de dados e retorna o ultimo registro de ponto do colaborador
func SelectPunch(nome string, dynamoClient dynamodb.Client, logs configuration.GoAppTools) model.Punch {
	input := &dynamodb.QueryInput{
		TableName:              aws.String("PontoColaborador"),
		ScanIndexForward:       aws.Bool(false),
		Limit:                  aws.Int32(1),
		KeyConditionExpression: aws.String("#nome = :n"),
		ExpressionAttributeNames: map[string]string{
			"#nome": "Nome",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":n": &types.AttributeValueMemberS{Value: nome},
		},
	}

	resp, err := dynamoClient.Query(context.Background(), input)
	configuration.Check(err, logs)
	if len(resp.Items) == 0 {
		err = errors.New("nao foi encontrado nenhum registro")
		configuration.Check(err, logs)
		return model.Punch{}
	}

	punch := model.Punch{
		Nome: resp.Items[0]["Nome"].(*types.AttributeValueMemberS).Value,
		Data: resp.Items[0]["Data"].(*types.AttributeValueMemberS).Value,
	}

	return punch
}

// SelectReport faz uma query no banco de dados e retorna todos os registros de ponto do colaborador no mes informado
func SelectReport(nome string, periodo string, dynamoClient dynamodb.Client, logs configuration.GoAppTools) []model.Punch {
	ctx := context.Background()

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String("PontoColaborador"),
		KeyConditionExpression: aws.String("#Nome = :nome AND begins_with(#Data, :mes)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":nome": &types.AttributeValueMemberS{Value: nome},
			":mes":  &types.AttributeValueMemberS{Value: periodo},
		},
		ExpressionAttributeNames: map[string]string{
			"#Nome": "Nome",
			"#Data": "Data",
		},
	}
	queryOutput, err := dynamoClient.Query(ctx, queryInput)
	configuration.Check(err, logs)

	punchs := make([]model.Punch, 0)
	for _, item := range queryOutput.Items {
		punch := model.Punch{}
		err := attributevalue.UnmarshalMap(item, &punch)
		if err != nil {
			// Handle the error
			configuration.Check(err, logs)
			continue
		}
		punchs = append(punchs, punch)
	}

	return punchs
}
