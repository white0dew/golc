package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/hupe1980/golc/embedding"
	"github.com/hupe1980/golc/vectorstore"
	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

var (
	addr                  = flag.String("addr", "aistar.cool:6334", "the address to connect to")
	collectionName        = "test_collection"
	vectorSize     uint64 = 4
	distance              = pb.Distance_Dot
)

func main1() {
	// Set up a connection to the server.
	conn, err := grpc.DialContext(context.Background(), *addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// create grpc collection client
	collections_client := pb.NewCollectionsClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*19)
	defer cancel()

	// Check Qdrant version
	qdrantClient := pb.NewQdrantClient(conn)
	healthCheckResult, err := qdrantClient.HealthCheck(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		log.Fatalf("Could not get health: %v", err)
	} else {
		log.Printf("Qdrant version: %s", healthCheckResult.GetVersion())
	}

	// Delete collection
	_, err = collections_client.Delete(ctx, &pb.DeleteCollection{
		CollectionName: collectionName,
	})
	if err != nil {
		log.Fatalf("Could not delete collection: %v", err)
	} else {
		log.Println("Collection", collectionName, "deleted")
	}

	// Create new collection
	var defaultSegmentNumber uint64 = 2
	_, err = collections_client.Create(ctx, &pb.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: &pb.VectorsConfig{Config: &pb.VectorsConfig_Params{
			Params: &pb.VectorParams{
				Size:     vectorSize,
				Distance: distance,
			},
		}},

		OptimizersConfig: &pb.OptimizersConfigDiff{
			DefaultSegmentNumber: &defaultSegmentNumber,
		},
	})
	if err != nil {
		log.Fatalf("Could not create collection: %v", err)
	} else {
		log.Println("Collection", collectionName, "created")
	}

	// List all created collections
	r, err := collections_client.List(ctx, &pb.ListCollectionsRequest{})
	if err != nil {
		log.Fatalf("Could not get collections: %v", err)
	} else {
		log.Printf("List of collections: %s", r.GetCollections())
	}
	// Create points grpc client
	pointsClient := pb.NewPointsClient(conn)

	// Create keyword field index
	fieldIndex1Type := pb.FieldType_FieldTypeKeyword
	fieldIndex1Name := "city"
	_, err = pointsClient.CreateFieldIndex(ctx, &pb.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      fieldIndex1Name,
		FieldType:      &fieldIndex1Type,
	})
	if err != nil {
		log.Fatalf("Could not create field index: %v", err)
	} else {
		log.Println("Field index for field", fieldIndex1Name, "created")
	}

	// Create integer field index
	fieldIndex2Type := pb.FieldType_FieldTypeInteger
	fieldIndex2Name := "count"
	_, err = pointsClient.CreateFieldIndex(ctx, &pb.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      fieldIndex2Name,
		FieldType:      &fieldIndex2Type,
	})
	if err != nil {
		log.Fatalf("Could not create field index: %v", err)
	} else {
		log.Println("Field index for field", fieldIndex2Name, "created")
	}

	// Upsert points
	waitUpsert := true
	upsertPoints := []*pb.PointStruct{
		{
			// Point Id is number or UUID
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 1},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.05, 0.61, 0.76, 0.74}}}},
			Payload: map[string]*pb.Value{
				"city": {
					Kind: &pb.Value_StringValue{StringValue: "Berlin"},
				},
				"country": {
					Kind: &pb.Value_StringValue{StringValue: "Germany"},
				},
				"count": {
					Kind: &pb.Value_IntegerValue{IntegerValue: 1000000},
				},
				"square": {
					Kind: &pb.Value_DoubleValue{DoubleValue: 12.5},
				},
			},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 2},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.19, 0.81, 0.75, 0.11}}}},
			Payload: map[string]*pb.Value{
				"city": {
					Kind: &pb.Value_ListValue{
						ListValue: &pb.ListValue{
							Values: []*pb.Value{
								{
									Kind: &pb.Value_StringValue{StringValue: "Berlin"},
								},
								{
									Kind: &pb.Value_StringValue{StringValue: "London"},
								},
							},
						},
					},
				},
			},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 3},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.36, 0.55, 0.47, 0.94}}}},
			Payload: map[string]*pb.Value{
				"city": {
					Kind: &pb.Value_ListValue{
						ListValue: &pb.ListValue{
							Values: []*pb.Value{
								{
									Kind: &pb.Value_StringValue{StringValue: "Berlin"},
								},
								{
									Kind: &pb.Value_StringValue{StringValue: "Moscow"},
								},
							},
						},
					},
				},
			},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 4},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.18, 0.01, 0.85, 0.80}}}},
			Payload: map[string]*pb.Value{
				"city": {
					Kind: &pb.Value_ListValue{
						ListValue: &pb.ListValue{
							Values: []*pb.Value{
								{
									Kind: &pb.Value_StringValue{StringValue: "London"},
								},
								{
									Kind: &pb.Value_StringValue{StringValue: "Moscow"},
								},
							},
						},
					},
				},
			},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 5},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.24, 0.18, 0.22, 0.44}}}},
			Payload: map[string]*pb.Value{
				"count": {
					Kind: &pb.Value_ListValue{
						ListValue: &pb.ListValue{
							Values: []*pb.Value{
								{
									Kind: &pb.Value_IntegerValue{IntegerValue: 0},
								},
							},
						},
					},
				},
			},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 6},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.35, 0.08, 0.11, 0.44}}}},
			Payload: map[string]*pb.Value{},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: "58384991-3295-4e21-b711-fd3b94fa73e3"},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.35, 0.08, 0.11, 0.44}}}},
			Payload: map[string]*pb.Value{},
		},
	}
	_, err = pointsClient.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &waitUpsert,
		Points:         upsertPoints,
	})
	if err != nil {
		log.Fatalf("Could not upsert points: %v", err)
	} else {
		log.Println("Upsert", len(upsertPoints), "points")
	}

	// Retrieve points by ids
	pointsById, err := pointsClient.Get(ctx, &pb.GetPoints{
		CollectionName: collectionName,
		Ids: []*pb.PointId{
			{PointIdOptions: &pb.PointId_Num{Num: 1}},
			{PointIdOptions: &pb.PointId_Num{Num: 2}},
		},
	})
	if err != nil {
		log.Fatalf("Could not retrieve points: %v", err)
	} else {
		log.Printf("Retrieved points: %s", pointsById.GetResult())
	}

	// Unfiltered search
	unfilteredSearchResult, err := pointsClient.Search(ctx, &pb.SearchPoints{
		CollectionName: collectionName,
		Vector:         []float32{0.2, 0.1, 0.9, 0.7},
		Limit:          3,
		// Include all payload and vectors in the search result
		WithVectors: &pb.WithVectorsSelector{SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: true}},
		WithPayload: &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		log.Fatalf("Could not search points: %v", err)
	} else {
		log.Printf("Found points: %s", unfilteredSearchResult.GetResult())
	}

	// filtered search
	filteredSearchResult, err := pointsClient.Search(ctx, &pb.SearchPoints{
		CollectionName: collectionName,
		Vector:         []float32{0.2, 0.1, 0.9, 0.7},
		Limit:          3,
		Filter: &pb.Filter{
			Should: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "city",
							Match: &pb.Match{
								MatchValue: &pb.Match_Keyword{
									Keyword: "London",
								},
							},
						},
					},
				},
			},
			MustNot: nil,
		},
		WithVectors: &pb.WithVectorsSelector{SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: true}},
		WithPayload: &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		log.Fatalf("Could not search points: %v", err)
	} else {
		log.Printf("Found points: %v", filteredSearchResult.GetTime())
	}
	return
}

func main() {
	openai, err := embedding.NewOpenAI("sk-peisUrRs7gPLZKPk3c758475E6604f87B427Df3f4f34Cd45", func(o *embedding.OpenAIOptions) {
		o.BaseURL = "https://35.nekoapi.com/v1"
	})
	if err != nil {
		log.Fatal(err)
	}

	qdrant, err := vectorstore.NewQdrant(openai, "text", func(options *vectorstore.QdrantOptions) {
		options.GroupId = 1
		options.TopK = 3
		options.AddrPort = "aistar.cool:6334"
		options.CollectionName = "aistar"
		options.ScoreThreshold = 0.1
	})
	//err = qdrant.AddDocuments(context.Background(), []schema.Document{{
	//	PageContent: "import requests\nimport chardet\nimport json\nurl = \"http://172.20.81.210:8000/api/chat\"\n\n\n\nheaders = {\n    \"Content-Type\": \"application/json\",\n    \"Authorization\": \"Bearer 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\"\n}\n\ndata = {\n    \"max_tokens\": 50,\n    \"stop\": [\"\\n\"],\n    \"temperature\": 0.5,\n    \"model\":\"gpt-3.5-turbo\",\n    \"messages\": [{\"role\": \"user\", \"content\": \"Hello!write 1000字论文呢\"}],\n    print(json.loads(event))\n\n# import gzip\n\n# # 原始字节字符串\n# compressed_data = b'\\x1f\\x8b\\x08\\x00\\x00\\x00\\x00\\x00\\x00\\x03T\\x8e\\xdbJ\\x031\\x14E\\xdf\\xfd\\x8a\\xb8\\x9f3e\\xc6j[\\xf3\"\\x14\\x84*\\x08\\x8a\\n^\\x90\\x92&\\xc7Nj&g\\x9c\\x9c\\xa2\\xb5\\xcc\\xbfK\\xf1\\x86\\xaf\\x8b\\xbd\\x17k\\x8b\\xe0a\\xe0j+\\xaeic1\\x1e\\xdd\\xf0\\xaa:\\xbd\\x7fX\\xc6\\xe1\\xeb\\xc5\\xdd\\xd5\\xed\\xe5\\xf4z\\xea?\\xc8\\xf3y\\x1d\\xa0\\xc1\\x8b\\x159\\xf9~\\x0c\\x1c7m$\\t\\x9c\\xa0\\xe1:\\xb2B\\x1e\\xa6\\x1aM\\xaa\\xf1xr|Xi4\\xec)\\xc2`\\xd9J1\\x1c\\x1c\\x15\\xb2\\xee\\x16\\\\\\x94\\xc3\\xb2\\x82\\xc6:\\xdb%\\xc1l\\xd1v\\xdc\\xb42\\x17~\\xa1\\x94a\\xaaR\\xe3O\\xfd\\x0f\\x0b\\x8b\\x8d\\xbf\\xe4\\xa0\\xec5\\\\\\xcd\\xc1Q\\x86y\\xdc\\xa2\\xa1\\xfc\\xe3\\xec8\\x12\\x0cl\\xce!\\x8bM\\xb2+\\xe4$\\x94v\\xf53\\x8a\\x91\\x95\\xd4\\xd4\\xd1\\xbe\\x9a\\xf1\\x9br6\\xa93\\xf55V\\x1b^+ao7\\'\\xe85\\x9eC\\n\\xb9\\x9ewd3\\'\\x18d\\xe1\\x16\\x1a!yz\\x87)\\xfb\\xa7~\\xef\\x13\\x00\\x00\\xff\\xff\\x03\\x00\\xa1\\x89W\\xd5F\\x01\\x00\\x00'# 解压缩字节数据并将其解码为文本\n# decoded_data = gzip.decompress(compressed_data).decode('utf-8')\n\n# print(decoded_data) map[title:test.py]}",
	//	Metadata:    nil,
	//}})
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	res, err := qdrant.SimilaritySearch(context.Background(), "白嫖云电脑注册教程")
	fmt.Println(res)
}
