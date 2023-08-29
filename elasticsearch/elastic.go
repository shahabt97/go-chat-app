package elasticsearch

import (
	"context"
	"io"
	"time"

	es "github.com/elastic/go-elasticsearch/v7"
)

var EsClient, Err = es.NewDefaultClient()

type ElasticClient struct {
	Client *es.Client
}

var Client = &ElasticClient{
	Client: EsClient,
}

func (client *ElasticClient) CreateDoc(Index string, body io.Reader) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.Client.Index(Index, body, client.Client.Index.WithContext(ctx))
	if err != nil {
		return err
	}
	return nil
}
