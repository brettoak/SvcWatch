package storage

import (
	"fmt"
	"strings"
)

// QueryLogs executes a dynamic query across one or multiple tables with pagination and filtering.
func (s *SqliteStorage) QueryLogs(tableNames []string, filter LogQueryFilter) (*LogQueryResponse, error) {
	if len(tableNames) == 0 {
		return &LogQueryResponse{Total: 0, Page: filter.Page, Size: filter.Size, Items: []LogQueryItem{}}, nil
	}

	// 1. Build the base FROM clause with UNION ALL
	var unionParts []string
	for _, table := range tableNames {
		unionParts = append(unionParts, fmt.Sprintf("SELECT '%s' as source_id, * FROM %s", table, table))
	}
	baseTable := fmt.Sprintf("(%s)", strings.Join(unionParts, " UNION ALL "))

	// 2. Build WHERE clauses
	var whereClauses []string
	var args []interface{}

	if filter.StartTime != "" {
		whereClauses = append(whereClauses, "time_local >= ?")
		args = append(args, filter.StartTime)
	}
	if filter.EndTime != "" {
		whereClauses = append(whereClauses, "time_local <= ?")
		args = append(args, filter.EndTime)
	}
	if filter.IP != "" {
		whereClauses = append(whereClauses, "remote_addr LIKE ?")
		args = append(args, filter.IP+"%")
	}
	if filter.Method != "" {
		whereClauses = append(whereClauses, "request LIKE ?")
		args = append(args, filter.Method+" %")
	}
	if filter.Status != nil {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, *filter.Status)
	} else if filter.StatusClass != "" {
		if len(filter.StatusClass) == 3 && filter.StatusClass[1:] == "xx" {
			char := filter.StatusClass[0]
			if char >= '1' && char <= '5' {
				minStatus := int(char-'0') * 100
				maxStatus := minStatus + 99
				whereClauses = append(whereClauses, "status >= ? AND status <= ?")
				args = append(args, minStatus, maxStatus)
			}
		}
	}
	if filter.PathKeyword != "" {
		whereClauses = append(whereClauses, "request LIKE ?")
		args = append(args, "%"+filter.PathKeyword+"%")
	}
	if filter.MinLatency != nil {
		whereClauses = append(whereClauses, "request_time >= ?")
		args = append(args, float64(*filter.MinLatency)/1000.0)
	}
	if filter.MaxLatency != nil {
		whereClauses = append(whereClauses, "request_time <= ?")
		args = append(args, float64(*filter.MaxLatency)/1000.0)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 3. Count Total
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", baseTable, whereSQL)
	var total int
	err := s.db.QueryRow(countSQL, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count logs: %w", err)
	}

	// 4. Query Items
	orderSQL := "ORDER BY time_local DESC"
	if filter.Sort == "latency_desc" {
		orderSQL = "ORDER BY request_time DESC"
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Size < 1 {
		filter.Size = 50
	} else if filter.Size > 500 {
		filter.Size = 500
	}

	limit := filter.Size
	offset := (filter.Page - 1) * filter.Size

	querySQL := fmt.Sprintf("SELECT source_id, id, remote_addr, remote_user, time_local, request, status, body_bytes_sent, http_referer, http_user_agent, request_time FROM %s %s %s LIMIT ? OFFSET ?", baseTable, whereSQL, orderSQL)
	
	queryArgs := append(args, limit, offset)
	
	rows, err := s.db.Query(querySQL, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var items []LogQueryItem
	for rows.Next() {
		var item LogQueryItem
		err := rows.Scan(
			&item.SourceID,
			&item.Entry.ID,
			&item.Entry.RemoteAddr,
			&item.Entry.RemoteUser,
			&item.Entry.TimeLocal,
			&item.Entry.Request,
			&item.Entry.Status,
			&item.Entry.BodyBytesSent,
			&item.Entry.HttpReferer,
			&item.Entry.HttpUserAgent,
			&item.Entry.RequestTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log row: %w", err)
		}
		items = append(items, item)
	}

	if items == nil {
		items = make([]LogQueryItem, 0)
	}

	return &LogQueryResponse{
		Total: total,
		Page:  filter.Page,
		Size:  filter.Size,
		Items: items,
	}, nil
}
