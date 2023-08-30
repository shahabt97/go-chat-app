package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	es "github.com/elastic/go-elasticsearch/v7"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (client *ElasticClient) SearchUser(q string, indexes ...string) ([]string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf(`{
	"query": {
		"query_string": {
	  		"fields": ["username","email"],
	  		"query": "*%s*"
				}
			}
  		}`, q)

	result, err := client.Client.Search(client.Client.Search.WithIndex(indexes...),
		client.Client.Search.WithContext(ctx),
		client.Client.Search.WithBody(strings.NewReader(query)))

	if err != nil {
		fmt.Println("Error in SearchContain function: ", err)
		return nil, err
	}
	// fmt.Println("result: ", result)

	res := make(map[string]interface{})

	errInDecode := json.NewDecoder(result.Body).Decode(&res)
	if errInDecode != nil {
		fmt.Println("error in decoding res.body of elastic search result: ", err)
		return nil, errInDecode
	}

	hitsRaw := res["hits"].(map[string]interface{})["hits"].([]interface{})
	var hits []string
	for _, hit := range hitsRaw {
		data := hit.(map[string]interface{})["_source"].(map[string]interface{})
		hits = append(hits, data["username"].(string))
	}

	return hits, nil
}

func (client *ElasticClient) SearchPubMessages(q string, indexes ...string) ([]primitive.ObjectID, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf(`	{
		"query": {
		"query_string": {
		  "fields": ["message"],
		  "query": "*%s*"
		}
		}
	  }`, q)

	result, err := client.Client.Search(client.Client.Search.WithIndex(indexes...),
		client.Client.Search.WithContext(ctx),
		client.Client.Search.WithBody(strings.NewReader(query)))

	if err != nil {
		fmt.Println("Error in searching pubMessages : ", err)
		return nil, err
	}
	// fmt.Println("result: ", result)

	res := make(map[string]interface{})

	errInDecode := json.NewDecoder(result.Body).Decode(&res)
	if errInDecode != nil {
		fmt.Println("error in decoding res.body of elastic search result: ", err)
		return nil, errInDecode
	}

	hitsRaw := res["hits"].(map[string]interface{})["hits"].([]interface{})
	var hits []primitive.ObjectID
	for _, hit := range hitsRaw {
		data := hit.(map[string]interface{})["_source"].(map[string]interface{})
		objectID, err := primitive.ObjectIDFromHex(data["id"].(string))
		if err != nil {
			return nil, err
		}
		hits = append(hits, objectID)
	}

	return hits, nil

}

func (client *ElasticClient) SearchPvMessages(q string, user string, host string, indexes ...string) ([]primitive.ObjectID, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf(`	{
		"query": {
		  "bool": {
			
			"should": [
			  {
				"bool": {
				  "must": [
							{  
						 "query_string": {
							  "fields": ["message"],
							  "query": "*%s*"
									  }
								  },
							{
							  "term": {
								"sender.keyword": {
								  "value": "%s"
								}
							  }
							},
							{
							  "term": {
								"receiver.keyword": {
								  "value": "%s"
								}
							  }
							}
						   ]
					   }
			  },
			  {
				"bool": {
				  "must": [
							{  
						 "query_string": {
							  "fields": ["message"],
							  "query": "*%s*"
									  }
								  },
							{
							  "term": {
								"sender.keyword": {
								  "value": "%s"
								}
							  }
							},
							{
							  "term": {
								"receiver.keyword": {
								  "value": "%s"
								}
							  }
							}
						   ]
					   }
			  }
			],
			
			"minimum_should_match": 1
		  }
		}
	  }`, q, user, host, q, host, user)

	result, err := client.Client.Search(client.Client.Search.WithIndex(indexes...),
		client.Client.Search.WithContext(ctx),
		client.Client.Search.WithBody(strings.NewReader(query)))

	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})

	errInDecode := json.NewDecoder(result.Body).Decode(&res)
	if errInDecode != nil {
		return nil, errInDecode
	}

	hitsRaw := res["hits"].(map[string]interface{})["hits"].([]interface{})
	var hits []primitive.ObjectID
	for _, hit := range hitsRaw {
		data := hit.(map[string]interface{})["_source"].(map[string]interface{})
		objectID, err := primitive.ObjectIDFromHex(data["id"].(string))
		if err != nil {
			return nil, err
		}
		hits = append(hits, objectID)
	}

	return hits, nil

}
