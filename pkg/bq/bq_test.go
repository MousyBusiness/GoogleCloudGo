package bq

import (
	"testing"
)

func TestBuildQuery(t *testing.T) {
	// single field
	query := BuildQuery("FROM `proj.dataset.table`", []string{"field_1"}, "WHERE field_1='123'")
	want := "SELECT field_1 FROM `proj.dataset.table` WHERE field_1='123'"
	if query != want {
		t.Errorf("Query incorrectly generated, wanted: %v, got: %v", want, query)
	}

	// multiple fields
	query = BuildQuery("FROM `proj.dataset.table`", []string{"field_1", "field_2"}, "WHERE field_1='123'")
	want = "SELECT field_1,field_2 FROM `proj.dataset.table` WHERE field_1='123'"
	if query != want {
		t.Errorf("Query incorrectly generated, wanted: %v, got: %v", want, query)
	}

	// empty fields
	query = BuildQuery("FROM `proj.dataset.table`", []string{}, "WHERE field_1='123'")
	want = "SELECT * FROM `proj.dataset.table` WHERE field_1='123'"
	if query != want {
		t.Errorf("Query incorrectly generated, wanted: %v, got: %v", want, query)
	}

	// nil fields
	query = BuildQuery("FROM `proj.dataset.table`", nil, "WHERE field_1='123'")
	want = "SELECT * FROM `proj.dataset.table` WHERE field_1='123'"
	if query != want {
		t.Errorf("Query incorrectly generated, wanted: %v, got: %v", want, query)
	}

	// end to end SELECT *
	query = BuildQuery(BuildFromClause("proj", "dataset", "table"), nil, BuildWhereClause("field_1", "123"))
	want = "SELECT * FROM `proj.dataset.table` WHERE field_1='123'"
	if query != want {
		t.Errorf("Query incorrectly generated, wanted: %v, got: %v", want, query)
	}

	// end to end SELECT field_1 FROM..
	query = BuildQuery(BuildFromClause("proj", "dataset", "table"), []string{"field_1"}, BuildWhereClause("field_1", "123"))
	want = "SELECT field_1 FROM `proj.dataset.table` WHERE field_1='123'"
	if query != want {
		t.Errorf("Query incorrectly generated, wanted: %v, got: %v", want, query)
	}

	// end to end SELECT field_1,field_2 FROM..
	query = BuildQuery(BuildFromClause("proj", "dataset", "table"), []string{"field_1", "field_2"}, BuildWhereClause("field_1", "123"))
	want = "SELECT field_1,field_2 FROM `proj.dataset.table` WHERE field_1='123'"
	if query != want {
		t.Errorf("Query incorrectly generated, wanted: %v, got: %v", want, query)
	}

	// end to end WHERE IN('123','234')
	query = BuildQuery(BuildFromClause("proj", "dataset", "table"), []string{"field_1", "field_2"}, BuildWhereClause("field_1", "123", "234"))
	want = "SELECT field_1,field_2 FROM `proj.dataset.table` WHERE field_1 IN('123','234')"
	if query != want {
		t.Errorf("Query incorrectly generated, wanted: %v, got: %v", want, query)
	}

	// end to end no WHERE
	query = BuildQuery(BuildFromClause("proj", "dataset", "table"), []string{"field_1", "field_2"}, BuildWhereClause("field_1"))
	want = "SELECT field_1,field_2 FROM `proj.dataset.table`"
	if query != want {
		t.Errorf("Query incorrectly generated, wanted: %v, got: %v", want, query)
	}
}

func TestBuildWhereClause(t *testing.T) {
	where := BuildWhereClause("field_1", "123", "234")
	want := `WHERE field_1 IN('123','234')`
	if where != want {
		t.Errorf("Where clause incorrectly generated, wanted: %v, got: %v", want, where)
	}

	// field is empty
	where = BuildWhereClause("", "123", "234")
	want = ""
	if where != "" {
		t.Errorf("Where clause incorrectly generated, wanted: %v, got: %v", want, where)
	}

	// values are empty
	where = BuildWhereClause("")
	want = ""
	if where != "" {
		t.Errorf("Where clause incorrectly generated, wanted: %v, got: %v", want, where)
	}

	// values are nil
	where = BuildWhereClause("field_1")
	want = ""
	if where != "" {
		t.Errorf("Where clause incorrectly generated, wanted: %v, got: %v", want, where)
	}
}

func TestBuildFromClause(t *testing.T) {
	from := BuildFromClause("proj", "dataset", "table")
	want := "FROM `proj.dataset.table`"
	if from != want {
		t.Errorf("From clause incorrectly generated, wanted: %v, got: %v", want, from)
	}
}
