package main

import (
    "fmt"
    "github.com/Pallinder/go-randomdata"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/graphql-go/graphql/language/ast"
    "math/rand"
    //"github.com/graph-gophers/graphql-go/relay"
    "github.com/graphql-go/graphql"
    "github.com/graphql-go/handler"
    "log"
)

var UUIDScalarType = graphql.NewScalar(graphql.ScalarConfig{
    Name:        "UUID",
    Description: "The `UUID` scalar type represents an ID Object.",
    Serialize: func(value interface{}) interface{} {
        switch value := value.(type) {
        case uuid.UUID:
            return value.String()
        case *uuid.UUID:
            v := *value
            return v.String()
        default:
            return nil
        }
    },
    ParseValue: func(value interface{}) interface{} {
        switch value := value.(type) {
        case string:
            return uuid.MustParse(value)
        case *string:
            return uuid.MustParse(*value)
        default:
            return nil
        }
    },
    ParseLiteral: func(valueAST ast.Value) interface{} {
        switch valueAST := valueAST.(type) {
        case *ast.StringValue:
            return uuid.MustParse(valueAST.Value)
        default:
            return nil
        }
    },
})

type Project struct {
    Hash         uuid.UUID     `json:"hash"`
    Title        string        `json:"title"`
    Description  string        `json:"description"`
    Parent       ProjectNoDeep `json:"parent,omitempty"`
    Repositories []Repository  `json:"repositories"`
}

type ProjectNoDeep struct {
    Hash  uuid.UUID `json:"hash"`
    Title string    `json:"title"`
}

var markdownType = graphql.NewObject(
    graphql.ObjectConfig{
        Name: "Markdown",
        Fields: graphql.Fields{
            "hash": &graphql.Field{
                Type: UUIDScalarType,
            },
            "file": &graphql.Field{
                Type: graphql.String,
            },
        },
    },
)

var repositoryType = graphql.NewObject(
    graphql.ObjectConfig{
        Name: "Repository",
        Fields: graphql.Fields{
            "hash": &graphql.Field{
                Type: UUIDScalarType,
            },
            "title": &graphql.Field{
                Type: graphql.String,
            },
            "description": &graphql.Field{
                Type: graphql.String,
            },
            "markdowns": &graphql.Field{
                Type: graphql.NewList(markdownType),
            },
        },
    },
)

var productNoDeepType = graphql.NewObject(
    graphql.ObjectConfig{
        Name: "ProjectParent",
        Fields: graphql.Fields{
            "hash": &graphql.Field{
                Type: UUIDScalarType,
            },
            "title": &graphql.Field{
                Type: graphql.String,
            },
        },
    },
)

var projectType = graphql.NewObject(
    graphql.ObjectConfig{
        Name: "Project",
        Fields: graphql.Fields{
            "hash": &graphql.Field{
                Type: UUIDScalarType,
            },
            "title": &graphql.Field{
                Type: graphql.String,
            },
            "description": &graphql.Field{
                Type: graphql.String,
            },
            "parent": &graphql.Field{
                Type: productNoDeepType,
            },
            "repositories": &graphql.Field{
                Type: graphql.NewList(repositoryType),
            },
        },
    },
)

type Repository struct {
    Hash        uuid.UUID  `json:"hash"`
    Title       string     `json:"title"`
    Description string     `json:"description"`
    Markdowns   []Markdown `json:"markdowns"`
}

type Markdown struct {
    Hash uuid.UUID `json:"hash"`
    File string    `json:"file"`
}

func FakeRepository() Repository {
    return Repository{
        Hash:        uuid.New(),
        Title:       randomdata.Adjective(),
        Description: randomdata.Paragraph(),
        //Markdowns:
    }
}

func FakeProjectNoDeep() ProjectNoDeep {
    return ProjectNoDeep{
        Hash:  uuid.New(),
        Title: randomdata.FirstName(rand.Intn(2)),
    }
}

func FakeProject() Project {
    preps := make([]Repository, rand.Intn(8)+4)
    for i := range preps {
        preps[i] = FakeRepository()
    }
    return Project{
        Hash:         uuid.New(),
        Title:        randomdata.FirstName(rand.Intn(2)),
        Description:  randomdata.Paragraph(),
        Parent:       FakeProjectNoDeep(),
        Repositories: preps,
    }
}

func main() {

    r := gin.Default()
    fmt.Println(randomdata.SillyName())

    // Schema
    fields := graphql.Fields{
        "project": &graphql.Field{
            Type: projectType,
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                return FakeProject(), nil
            },
        },
        "projects": &graphql.Field{
            Type: graphql.NewList(projectType),
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                arr := make([]Project, rand.Intn(8)+4)
                for i := range arr {
                    arr[i] = FakeProject()
                }
                return arr, nil
            },
        },
        "repositories": &graphql.Field{
            Type: graphql.NewList(repositoryType),
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                arr := make([]Repository, rand.Intn(8)+4)
                for i := range arr {
                    arr[i] = FakeRepository()
                }
                return arr, nil
            },
        },
    }
    rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
    schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
    schema, err := graphql.NewSchema(schemaConfig)
    if err != nil {
        log.Fatalf("failed to create new schema, error: %v", err)
    }

    r.POST("/graphql", gin.WrapH(handler.New(&handler.Config{
        Schema:     &schema,
        Pretty:     true,
        GraphiQL:   true,
        Playground: false,
    })))

    r.Run()
}
