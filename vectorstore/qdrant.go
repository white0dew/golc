package vectorstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/hupe1980/golc/schema"
	"github.com/hupe1980/golc/util"
	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

var (
	KeyGroup = "group_id"
)

// Compile time check to ensure Qdrant_Old satisfies the VectorStore interface.
var _ schema.VectorStore = (*Qdrant)(nil)

type QdrantOptions struct {
	Namespace      string
	TopK           uint64
	AddrPort       string
	CollectionName string
	VectorSize     int
	GroupId        int
	ScoreThreshold float32
	KeyList        []string
}

type Qdrant struct {
	embedder schema.Embedder
	textKey  string
	opts     QdrantOptions
}

func NewQdrant(embedder schema.Embedder, textKey string, optFns ...func(*QdrantOptions)) (*Qdrant, error) {
	opts := QdrantOptions{
		TopK: 4,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	return &Qdrant{
		embedder: embedder,
		textKey:  textKey,
		opts:     opts,
	}, nil
}

func (q Qdrant) AddDocuments(ctx context.Context, docs []schema.Document) error {
	// 拿到groupId
	if groupId := q.opts.GroupId; groupId == 0 {
		return errors.New("groupId is empty")
	}

	// Set up a connection to the server.
	conn, err := grpc.DialContext(context.Background(), q.opts.AddrPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Check Qdrant_Old version
	qdrantClient := pb.NewQdrantClient(conn)
	_, err = qdrantClient.HealthCheck(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		return err
	}

	// 转换矢量
	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.PageContent
	}
	vectors, err := q.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}
	// 转换矢量
	points := q.schemaDocsToQdrantDocs(vectors, docs)
	// 更新
	waitUpsert := true
	pointsClient := pb.NewPointsClient(conn)
	_, err = pointsClient.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: q.opts.CollectionName,
		Wait:           &waitUpsert,
		Points:         points,
	})

	return err
}

func (q Qdrant) SimilaritySearch(ctx context.Context, query string) ([]schema.Document, error) {
	vector, err := q.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	//fmt.Println(vector)
	newVector := util.Map(vector, func(e float64, i int) float32 {
		return float32(e)
	})
	//fmt.Println(newVector)
	// 检查groupId
	if q.opts.GroupId == 0 {
		return nil, errors.New("group is empty")
	}

	// Set up a connection to the server.
	conn, err := grpc.DialContext(context.Background(), q.opts.AddrPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// Create points grpc client
	pointsClient := pb.NewPointsClient(conn)

	// filtered search
	searchResult, err := pointsClient.Search(ctx, &pb.SearchPoints{
		CollectionName: q.opts.CollectionName,
		Vector:         newVector,
		Filter: &pb.Filter{
			Must: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: KeyGroup,
							Match: &pb.Match{
								MatchValue: &pb.Match_Integer{
									Integer: int64(q.opts.GroupId),
								},
							},
						},
					},
				},
			},
		},
		Limit:          q.opts.TopK,
		WithPayload:    &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
		ScoreThreshold: &q.opts.ScoreThreshold,
	})

	if searchResult == nil || searchResult.Result == nil {
		fmt.Println(searchResult.GetTime())
		return nil, nil
	}
	fmt.Println(searchResult.GetTime())
	fmt.Println(searchResult.GetResult())
	return q.toSchemaDocs(searchResult), nil
}

func (q Qdrant) toSchemaDocs(searchResult *pb.SearchResponse) []schema.Document {
	docs := make([]schema.Document, len(searchResult.GetResult()))

	for k, v := range searchResult.GetResult() {
		pageContent := v.GetPayload()[q.textKey].GetStringValue()

		doc := schema.Document{
			PageContent: pageContent,
			//Metadata:    v.GetPayload(),
		}
		docs[k] = doc
	}

	return docs
}

func (q Qdrant) toPayLoad(doc schema.Document) map[string]*pb.Value {
	newMap := make(map[string]*pb.Value)

	// 原文
	newMap[q.textKey] = &pb.Value{
		Kind: &pb.Value_StringValue{StringValue: doc.PageContent},
	}

	// 归属权
	newMap[KeyGroup] = &pb.Value{
		Kind: &pb.Value_IntegerValue{IntegerValue: int64(q.opts.GroupId)},
	}

	// 批量key处理--暂且只处理string-TODO
	for _, v := range q.opts.KeyList {
		if str, ok := doc.Metadata[v].(string); ok {
			newMap[v] = &pb.Value{
				Kind: &pb.Value_StringValue{StringValue: str},
			}
		}
	}

	return newMap
}

func (q Qdrant) schemaDocsToQdrantDocs(vectors [][]float64, docs []schema.Document) []*pb.PointStruct {
	pbs := make([]*pb.PointStruct, len(docs))

	for i := 0; i < len(vectors); i++ {
		pb := &pb.PointStruct{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: uuid.New().String()},
			},
			Payload: q.toPayLoad(docs[i]),
			Vectors: &pb.Vectors{
				VectorsOptions: &pb.Vectors_Vector{
					Vector: &pb.Vector{
						Data: util.Map(vectors[i], func(e float64, i int) float32 {
							return float32(e)
						}),
					},
				},
			},
		}

		pbs[i] = pb
	}
	return pbs
}
