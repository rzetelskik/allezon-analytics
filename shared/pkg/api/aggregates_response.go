package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type Aggregate int

const (
	AGGREGATE_SUM_PRICE Aggregate = iota + 1
	AGGREGATE_COUNT
)

var stringToAggregate = map[string]Aggregate{
	"SUM_PRICE": AGGREGATE_SUM_PRICE,
	"COUNT":     AGGREGATE_COUNT,
}

func ParseAggregate(s string) (Aggregate, error) {
	a, ok := stringToAggregate[s]
	if !ok {
		return Aggregate(0), fmt.Errorf("%q is not a valid aggregate")
	}

	return a, nil
}

var aggregateToAggregateColumns = map[Aggregate]AggregateColumn{
	AGGREGATE_SUM_PRICE: SUM_PRICE,
	AGGREGATE_COUNT:     COUNT,
}

func AggregateToAggregateColumn(a Aggregate) AggregateColumn {
	return aggregateToAggregateColumns[a]
}

type AggregateColumn int

const (
	BUCKET AggregateColumn = iota + 1
	ACTION
	ORIGIN
	BRAND_ID
	CATEGORY_ID
	SUM_PRICE
	COUNT
)

//var columnToIndex = map[string]int{
//	"1m_bucket":   0,
//	"action":      1,
//	"origin":      2,
//	"brand_id":    3,
//	"category_id": 4,
//	"sum_price":   5,
//	"count":       6,
//}

var aggregateColumnToString = map[AggregateColumn]string{
	BUCKET:      "1m_bucket",
	ACTION:      "action",
	ORIGIN:      "origin",
	BRAND_ID:    "brand_id",
	CATEGORY_ID: "category_id",
	SUM_PRICE:   "sum_price",
	COUNT:       "count",
}

var aggregateColumnToIndex = map[AggregateColumn]int{
	BUCKET:      0,
	ACTION:      1,
	ORIGIN:      2,
	BRAND_ID:    3,
	CATEGORY_ID: 4,
	SUM_PRICE:   5,
	COUNT:       6,
}

func (ac AggregateColumn) string() string {
	return aggregateColumnToString[ac]
}

func (ac AggregateColumn) MarshalJSON() ([]byte, error) {
	return json.Marshal(ac.string())
}

type AggregateResponse struct {
	Columns []AggregateColumn `json:"columns"`
	Rows    []AggregateRow    `json:"-"`
}

type BucketTime time.Time

func (bt BucketTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(bt).Format("2006-01-02T15:04:05"))
}

type AggregateRow struct {
	Bucket     BucketTime
	Action     Action
	Origin     string // FIXME
	BrandID    string
	CategoryID string
	SumPrice   int64
	Count      int64
}

func (ar AggregateResponse) MarshalJSON() ([]byte, error) {
	type Alias AggregateResponse
	aux := &struct {
		*Alias
		Rows [][]interface{} `json:"rows"`
	}{
		Alias: (*Alias)(&ar),
		Rows:  make([][]interface{}, len(ar.Rows)),
	}

	for ir, r := range ar.Rows {
		data := reflect.ValueOf(r)
		dataArray := make([]interface{}, len(ar.Columns))
		for ic, c := range ar.Columns {
			dataArray[ic] = data.Field(aggregateColumnToIndex[c]).Interface()
		}
		aux.Rows[ir] = dataArray
	}

	return json.Marshal(aux)
}
