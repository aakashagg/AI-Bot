package data

import (
	"ai-bot/internal/types"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"log"
	"os"
)

const defaultRegion = "us-east-1"

type ThreadRepo struct {
	table dynamo.Table
}

func tableNotInList(tables []string, tableName string) bool {
	for i := range tables {
		if tables[i] == tableName {
			return false
		}
	}

	return true
}

func NewThreadRepo() *ThreadRepo {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = defaultRegion
	}
	sess := session.Must(session.NewSession())
	db := dynamo.New(sess, &aws.Config{Region: aws.String(region)})
	tableList := db.ListTables()
	tables, err := tableList.All()
	if err != nil {
		panic(err)
	}
	if tableNotInList(tables, "ai-bot") {
		err := db.CreateTable("ai-bot", types.Thread{}).OnDemand(true).Run()
		if err != nil {
			panic(err)
		}
	}

	table := db.Table("ai-bot")

	return &ThreadRepo{table: table}
}

func (r ThreadRepo) GetThread(ctx context.Context, threadTimestamp string) types.Thread {
	if threadTimestamp == "" {
		return types.Thread{}
	}

	var thread types.Thread

	err := r.table.Get("timestamp", threadTimestamp).OneWithContext(ctx, &thread)
	if err != nil {
		log.Println("Unable to find history of thread")
		return types.Thread{}
	}

	return thread
}

func (r ThreadRepo) SaveThread(ctx context.Context, thread types.Thread) error {
	return r.table.Put(thread).RunWithContext(ctx)
}
