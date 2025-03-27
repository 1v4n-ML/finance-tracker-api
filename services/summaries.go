package services

import (
	"fmt"
	"strings"

	"github.com/1v4n-ML/finance-tracker-api/models"
	"github.com/1v4n-ML/finance-tracker-api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func BuildAggregationPipeline(req models.AggregationRequest) (mongo.Pipeline, error) {
	pipeline := mongo.Pipeline{}

	// 1. $match stage (Filters)
	matchStage := bson.D{}
	if len(req.Filters) > 0 {
		filters := bson.D{}
		for _, f := range req.Filters {
			mongoOp, err := utils.MapOperator(f.Operator)
			if err != nil {
				return nil, fmt.Errorf("filter error: %w", err)
			}

			parsedValue, err := utils.ParseFilterValue(f.Field, f.Value, f.Operator)
			if err != nil {
				return nil, fmt.Errorf("filter value error: %w", err)
			}

			// Append filter condition. If the same field appears multiple times (e.g., date range),
			// MongoDB implicitly handles it with $and.
			filters = append(filters, bson.E{Key: f.Field, Value: bson.D{{Key: mongoOp, Value: parsedValue}}})
		}
		matchStage = bson.D{{Key: "$match", Value: filters}}
		pipeline = append(pipeline, matchStage)
	}

	// 2. $group stage (Grouping and Metrics)
	groupStage := bson.D{}
	groupID := bson.D{}      // _id field for grouping
	groupMetrics := bson.D{} // Accumulators for metrics

	// Build the _id part for grouping
	datePartsAdded := bson.M{} // Keep track of which date parts we've extracted
	for _, groupByField := range req.GroupBy {
		if strings.HasPrefix(groupByField, "date:") {
			part := strings.TrimPrefix(groupByField, "date:")
			fieldNameInID := part // e.g., "year", "month", "day"
			var dateOperator bson.D
			switch part {
			case "year":
				dateOperator = bson.D{{Key: "$year", Value: "$date"}}
			case "month":
				dateOperator = bson.D{{Key: "$month", Value: "$date"}}
			case "day":
				dateOperator = bson.D{{Key: "$dayOfMonth", Value: "$date"}}
			// Add more granularities if needed: week, dayOfYear, dayOfWeek
			default:
				return nil, fmt.Errorf("unsupported date grouping part: %s", part)
			}
			// Avoid duplicate date part extraction if multiple requested (e.g., year and month)
			if _, exists := datePartsAdded[part]; !exists {
				groupID = append(groupID, bson.E{Key: fieldNameInID, Value: dateOperator})
				datePartsAdded[part] = true
			}
		} else {
			// Use the original field name for the key in the _id document
			// Ensure the field name is valid in MongoDB (no dots, no dollar signs unless first char)
			safeFieldName := strings.ReplaceAll(groupByField, ".", "_") // Basic sanitization example
			groupID = append(groupID, bson.E{Key: safeFieldName, Value: "$" + groupByField})
		}
	}
	groupStage = append(groupStage, bson.E{Key: "_id", Value: groupID})

	// Build the metrics part (accumulators)
	for _, m := range req.Metrics {
		var accumulator bson.D
		switch m.Operation {
		case "sum":
			if m.Field == "" {
				return nil, fmt.Errorf("metric '%s': 'field' is required for 'sum' operation", m.Name)
			}
			accumulator = bson.D{{Key: "$sum", Value: "$" + m.Field}}
		case "count":
			accumulator = bson.D{{Key: "$sum", Value: 1}} // Count documents in the group
		case "avg":
			if m.Field == "" {
				return nil, fmt.Errorf("metric '%s': 'field' is required for 'avg' operation", m.Name)
			}
			accumulator = bson.D{{Key: "$avg", Value: "$" + m.Field}}
		default:
			return nil, fmt.Errorf("unsupported metric operation: %s", m.Operation)
		}
		groupMetrics = append(groupMetrics, bson.E{Key: m.Name, Value: accumulator})
	}
	groupStage = append(groupStage, groupMetrics...) // Append all metric accumulators
	pipeline = append(pipeline, bson.D{{Key: "$group", Value: groupStage}})

	// 3. $project stage (Optional - Reshape results)
	// Move grouped fields (_id.*) to top level for cleaner output
	projectStage := bson.D{}
	projectStage = append(projectStage, bson.E{Key: "_id", Value: 0}) // Exclude the original complex _id

	// Add grouped fields back at the top level
	for _, groupKey := range groupID {
		projectStage = append(projectStage, bson.E{Key: groupKey.Key, Value: "$_id." + groupKey.Key})
	}
	// Include calculated metrics
	for _, metricKey := range groupMetrics {
		projectStage = append(projectStage, bson.E{Key: metricKey.Key, Value: "$" + metricKey.Key})
	}
	pipeline = append(pipeline, bson.D{{Key: "$project", Value: projectStage}})

	// 4. $sort stage (Optional)
	if len(req.SortBy) > 0 {
		sortStage := bson.D{}
		for field, order := range req.SortBy {
			if order != 1 && order != -1 {
				return nil, fmt.Errorf("invalid sort order for field '%s': must be 1 or -1", field)
			}
			// Ensure the sort field exists in the $project stage output
			fieldExists := false
			for _, projKey := range projectStage {
				if projKey.Key == field {
					fieldExists = true
					break
				}
			}
			if !fieldExists {
				return nil, fmt.Errorf("cannot sort by field '%s': field not present in aggregation result", field)
			}
			sortStage = append(sortStage, bson.E{Key: field, Value: order})
		}
		pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sortStage}})
	}

	return pipeline, nil
}
