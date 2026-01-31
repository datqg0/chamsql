package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"judge/database"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

func sortResult(r [][]string) {
    sort.Slice(r, func(i, j int) bool {
        return fmt.Sprint(r[i]) < fmt.Sprint(r[j])
    })
}

func queryToSlice(db *sql.DB, query string) ([][]string, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	cols, _ := rows.Columns()
	raw := make([]sql.RawBytes, len(cols))
	scanArgs := make([]interface{}, len(raw))
	for i := range raw {
		scanArgs[i] = &raw[i]
	}

	var result [][]string
	for rows.Next() {
		rows.Scan(scanArgs...)
		var row []string
		for _, col := range raw {
			row = append(row, string(col))
		}
		result = append(result, row)
	}
	return result, nil
}

func JudgeSQL(c *gin.Context) {
	var req struct {
		ProblemID int    `json:"problem_id"`
		SQL       string `json:"sql"`
	}
	json.NewDecoder(c.Request.Body).Decode(&req)

	rows, err := database.JudgeDB.Query(
		"SELECT id, schema_sql, expected_sql FROM test_cases WHERE problem_id=?",
		req.ProblemID,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	total := 0
	pass := 0
	var errors []string

	for rows.Next() {
		total++

		var id int
		var schema, expected string
		if err := rows.Scan(&id, &schema, &expected); err != nil {
			errors = append(errors, fmt.Sprintf("Scan error: %v", err))
			continue
		}

		dbname := fmt.Sprintf("sandbox_%d_%d", req.ProblemID, id)
		sdb, err := database.CreateSandboxDB(dbname)
		if err != nil {
			errors = append(errors, fmt.Sprintf("CreateDB error: %v", err))
			continue
		}
		defer database.DropSandboxDB(dbname)

		sdb.Exec("SET FOREIGN_KEY_CHECKS = 0;")
		sdb.Exec("DROP TABLE IF EXISTS users, customers, orders, employees;")
		sdb.Exec("SET FOREIGN_KEY_CHECKS = 1;")

		// Execute schema statements separately
		schemaOk := true
		stmts := strings.Split(schema, ";")
		for _, stmt := range stmts {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := sdb.Exec(stmt); err != nil {
				sdb.Close()
				errors = append(errors, fmt.Sprintf("Schema exec error: %v (stmt: %s)", err, stmt))
				schemaOk = false
				break
			}
		}

		if !schemaOk {
			continue
		}

		exp, err1 := queryToSlice(sdb, expected)
		act, err2 := queryToSlice(sdb, req.SQL)

		if err1 != nil || err2 != nil {
			errors = append(errors, fmt.Sprintf("Query error - exp: %v, act: %v", err1, err2))
			sdb.Close()
			continue
		}

		sortResult(exp)
		sortResult(act)
		if reflect.DeepEqual(exp, act) {
			pass++
		} else {
			errors = append(errors, fmt.Sprintf("Test %d mismatch: expected %v, got %v", id, exp, act))
		}

		sdb.Close()
	}

	result := gin.H{"passed": pass, "total": total}
	if len(errors) > 0 {
		result["details"] = errors
	}
	c.JSON(http.StatusOK, result)
}
