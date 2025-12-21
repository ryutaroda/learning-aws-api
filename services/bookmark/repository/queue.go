package repository

import (
    "context"
    "encoding/json"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "bookmark/model"
)

type QueueRepository struct {
    sqsClient *sqs.Client
    queueURL  string
}

func NewQueueRepository(sqsClient *sqs.Client, queueURL string) *QueueRepository {
    return &QueueRepository{
        sqsClient: sqsClient,
        queueURL:  queueURL,
    }
}

func (r *QueueRepository) SendBookmarkCreated(ctx context.Context, bookmarkID uint, url string) error {
    msg := model.BookmarkCreatedMessage{
        BookmarkID: bookmarkID,
        URL:        url,
    }

    body, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    bodyStr := string(body)
    _, err = r.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
        QueueUrl:    &r.queueURL,
        MessageBody: &bodyStr,
    })

    return err
}
