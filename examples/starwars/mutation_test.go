package starwars_test

import (
	"reflect"
	"testing"

	"github.com/evenco/go-graphql"
	"github.com/evenco/go-graphql-relay/examples/starwars"
	"github.com/evenco/go-graphql/testutil"
)

func TestMutation_CorrectlyMutatesTheDataSet(t *testing.T) {
	query := `
      mutation AddBWingQuery($input: IntroduceShipInput!) {
        introduceShip(input: $input) {
          ship {
            id
            name
          }
          faction {
            name
          }
          clientMutationID
        }
      }
    `
	params := map[string]interface{}{
		"input": map[string]interface{}{
			"shipName":         "B-Wing",
			"factionId":        "1",
			"clientMutationID": "abcde",
		},
	}
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"introduceShip": map[string]interface{}{
				"ship": map[string]interface{}{
					"id":   "U2hpcDoxMA==",
					"name": "B-Wing",
				},
				"faction": map[string]interface{}{
					"name": "Alliance to Restore the Republic",
				},
				"clientMutationID": "abcde",
			},
		},
	}
	result := graphql.Graphql(graphql.Params{
		Schema:         starwars.Schema,
		RequestString:  query,
		VariableValues: params,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
