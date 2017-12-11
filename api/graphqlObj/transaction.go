package graphqlObj

import (
	"github.com/graphql-go/graphql"
)

var TransactionType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Transaction",
		Fields: graphql.Fields{
			"Id": &graphql.Field{
				Type: graphql.String,
			},
			"Type": &graphql.Field{
				Type: graphql.String,
			},
			"Verified": &graphql.Field{
				Type: graphql.Boolean,
			},
			"Transmitted": &graphql.Field{
				Type: graphql.Boolean,
			},
			"Properties": &graphql.Field{
				Type: PropertiesType,
			},
		},
	},
)
