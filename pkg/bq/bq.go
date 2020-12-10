package bq

import (
	"cloud.google.com/go/bigquery"
	"context"
	errs "github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"log"
	"os"
	"strings"
)

var (
	dataProjectId = os.Getenv("PROJECT")
	datasetName   = os.Getenv("DATASET_NAME")
	tableName     = os.Getenv("TABLE_NAME")
)

func QueryBQ(query string) ([]string, error) {
	log.Println("querying BigQuery with:", query)

	// check query is sane
	if !strings.HasPrefix(query, "SELECT") {
		log.Fatalln("invalid query command, malformed select!", query)
	}
	if !strings.Contains(query, "FROM ") {
		log.Fatalln("invalid query command, no from!", query)
	}

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, dataProjectId)
	if err != nil {
		return nil, errs.Wrap(err, "error fetching from BQ")
	}

	// perform query
	q := client.Query(query)

	// read data
	it, err := q.Read(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "error reading BQ dataset")
	}

	var rows [][]bigquery.Value
	for {
		var values []bigquery.Value
		err := it.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
			log.Println("error during BQ dataset iteration")
		}
		rows = append(rows, values)
	}

	log.Println(len(rows), "items returned from BQ")

	out := []string{}
	// convert []bigquery.Value to []string - there has to be a better way...
	for _, v := range rows {
		var s []string
		for _, vv := range v {
			s = append(s, vv.(string))
		}

		out = append(out, strings.Join(s, ","))
	}
	return out, nil
}

// returned 'SELECT *' when nil or empty columns are provided
func BuildQuery(from string, cols []string, whereClause string) string {
	sc := "SELECT "
	l := len(cols)

	if cols != nil && l != 0 {
		for i, v := range cols {
			sc = sc + v
			if i != l-1 {
				sc = sc + ","
			}
		}
	} else {
		// select all
		sc = sc + "*"
	}

	q := sc + " " + from
	if whereClause != "" {
		q = q + " " + whereClause
	}
	return q
}

func DefaultFrom() string {
	return BuildFromClause(dataProjectId, datasetName, tableName)
}

func BuildFromClause(proj string, dataset string, table string) string {
	if proj == "" || dataset == "" || table == "" {
		log.Fatalln("invalid table arguments! check environment variables are correctly deployed")
	}
	from := "FROM `" + proj + "." + dataset + "." + table + "`"
	return from
}

// return '' (empty where clause) if cols are empty, ids are nil or empty
func BuildWhereClause(col string, vals ...string) string {
	if col == "" || vals == nil || len(vals) == 0 {
		return ""
	}

	l := len(vals)
	c := "WHERE "
	if l > 1 {
		c = c + col + " IN("
		for i, v := range vals {
			c = c + "'" + v + "'"
			if i != l-1 {
				c = c + ","
			}
		}
		c = c + ")"
	} else {
		c = c + col + "='" + vals[0] + "'"
	}
	return c
}
