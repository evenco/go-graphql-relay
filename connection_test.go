package relay_test

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"
	"github.com/graphql-go/relay"
)

var connectionTestAllUsers = []interface{}{
	&user{Name: "Dan"},
	&user{Name: "Nick"},
	&user{Name: "Lee"},
	&user{Name: "Joe"},
	&user{Name: "Tim"},
}
var connectionTestUserType *graphql.Object
var connectionTestQueryType *graphql.Object
var connectionTestSchema graphql.Schema
var connectionTestConnectionDef *relay.GraphQLConnectionDefinitions

func init() {
	connectionTestUserType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.FieldConfigMap{
			"name": &graphql.FieldConfig{
				Type: graphql.String,
			},
			// re-define `friends` field later because `connectionTestUserType` has `connectionTestConnectionDef` has `connectionTestUserType` (cyclic-reference)
			"friends": &graphql.FieldConfig{},
		},
	})

	connectionTestConnectionDef = relay.ConnectionDefinitions(relay.ConnectionConfig{
		Name:     "Friend",
		NodeType: connectionTestUserType,
		EdgeFields: graphql.FieldConfigMap{
			"friendshipTime": &graphql.FieldConfig{
				Type: graphql.String,
				Resolve: func(ctx context.Context, p graphql.GQLFRParams) interface{} {
					return "Yesterday"
				},
			},
		},
		ConnectionFields: graphql.FieldConfigMap{
			"totalCount": &graphql.FieldConfig{
				Type: graphql.Int,
				Resolve: func(ctx context.Context, p graphql.GQLFRParams) interface{} {
					return len(connectionTestAllUsers)
				},
			},
		},
	})

	// define `friends` field here after getting connection definition
	connectionTestUserType.AddFieldConfig("friends", &graphql.FieldConfig{
		Type: connectionTestConnectionDef.ConnectionType,
		Args: relay.ConnectionArgs,
		Resolve: func(ctx context.Context, p graphql.GQLFRParams) interface{} {
			arg := relay.NewConnectionArguments(p.Args)
			res := relay.ConnectionFromArray(connectionTestAllUsers, arg)
			return res
		},
	})

	connectionTestQueryType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.FieldConfigMap{
			"user": &graphql.FieldConfig{
				Type: connectionTestUserType,
				Resolve: func(ctx context.Context, p graphql.GQLFRParams) interface{} {
					return connectionTestAllUsers[0]
				},
			},
		},
	})
	var err error
	connectionTestSchema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: connectionTestQueryType,
	})
	if err != nil {
		panic(err)
	}

}

func TestConnectionDefinition_IncludesConnectionAndEdgeFields(t *testing.T) {
	query := `
      query FriendsQuery {
        user {
          friends(first: 2) {
            totalCount
            edges {
              friendshipTime
              node {
                name
              }
            }
          }
        }
      }
    `
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"friends": map[string]interface{}{
					"totalCount": 5,
					"edges": []interface{}{
						map[string]interface{}{
							"friendshipTime": "Yesterday",
							"node": map[string]interface{}{
								"name": "Dan",
							},
						},
						map[string]interface{}{
							"friendshipTime": "Yesterday",
							"node": map[string]interface{}{
								"name": "Nick",
							},
						},
					},
				},
			},
		},
	}
	result := graphql.Graphql(graphql.Params{
		Schema:        connectionTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
