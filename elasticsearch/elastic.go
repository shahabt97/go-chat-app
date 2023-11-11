package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"go-chat-app/config"
	"io"
	"strings"
	"time"

	es "github.com/elastic/go-elasticsearch/v7"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var EsClient *es.Client

type ElasticClient struct {
	Client *es.Client
}

var Client *ElasticClient

func Init() (err error) {

	fmt.Printf("cfg is: %+v\n", config.ConfigData)

	EsClient, err = es.NewClient(es.Config{
		Addresses: []string{config.ConfigData.ElasticURI},
		Password:  config.ConfigData.ElasticPassword,
		Username:  config.ConfigData.ElasticUsername})

	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := EsClient.Ping(EsClient.Ping.WithContext(ctx))
	if err != nil {
		return
	}

	defer res.Body.Close()
		// fmt.Printf("res: %s\n", res.String())

	// Check the status code
	if res.IsError() {
		// The ping request failed
		// fmt.Printf("Error: %s\n", res.String())
		return fmt.Errorf(res.String())

	}

	Client = &ElasticClient{
		Client: EsClient,
	}
	return

}

// for creating a doc in elastic
func (client *ElasticClient) CreateDoc(Index string, body io.Reader) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Client.Index(Index, body, client.Client.Index.WithContext(ctx))
	if err != nil {
		return err
	}
	return nil
}

// search user based on username or email but in the end return username only
// username or email must contain search query
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
		fmt.Println("error in SearchContain function: ", err)
		return nil, err
	}

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

// search in all public or pv messages for data insights
// this function can violate user privacy so must be used carefully
func (client *ElasticClient) SearchAllMessages(q string, indexes ...string) ([]*AllMessages, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf(`	{
		"query": {
		  "match": {
			"message": "%s"
		  }
		}
	  }`, q)
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
	var hits []*AllMessages
	for _, hit := range hitsRaw {

		index := hit.(map[string]interface{})["_index"].(string)
		data := hit.(map[string]interface{})["_source"].(map[string]interface{})

		objectID, err := primitive.ObjectIDFromHex(data["id"].(string))
		if err != nil {
			return nil, err
		}

		hits = append(hits, &AllMessages{
			Index: index,
			Id:    objectID,
		})
	}

	return hits, nil

}
