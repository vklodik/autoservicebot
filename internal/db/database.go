package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbPath = "data/sqlite/sqlite.db"
)

var db *sql.DB

func Init() {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}

	q := `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, name TEXT, phone_number INTEGER);
		CREATE TABLE IF NOT EXISTS records (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, datetime INTEGER)`
	_, err = db.Exec(q)
	if err != nil {
		log.Printf("failed to create table: %s", err)
	}
}

func GetInfo(userId int) (string, int, error) {
	q := `SELECT name, phone_number FROM users WHERE user_id=?`

	var (
		name   string
		number int
	)
	err := db.QueryRow(q, userId).Scan(&name, &number)
	if err != nil {
		fmt.Printf("failed to check if row exists: %v", err)
		return "", 0, err
	}

	return name, number, nil
}

func SaveUser(userId int, name string, number int) error {
	q := `INSERT INTO users (user_id, name, phone_number) VALUES (?, ?, ?)`

	_, err := db.Exec(q, userId, name, number)
	if err != nil {
		return fmt.Errorf("failed to save data: %w", err)
	}

	return nil
}

func SaveRecord(userId int64, datetime int64) error {
	q := `INSERT INTO records (user_id, datetime) VALUES (?, ?)`

	_, err := db.Exec(q, userId, datetime)
	if err != nil {
		return fmt.Errorf("failed to save data: %w", err)
	}

	return nil
}

func GetAllRecords(userId int64) ([]int, error) {
	q := `SELECT datetime FROM records WHERE user_id=? AND datetime>strftime('%s', 'now') ORDER BY datetime`

	rows, err := db.Query(q, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	defer rows.Close()

	var datetimes []int
	for rows.Next() {
		var datetime int
		err = rows.Scan(&datetime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		datetimes = append(datetimes, datetime)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}

	return datetimes, nil
}

func GetAllTimes(result int64) ([]string, error) {
	q := `WITH available_times(available_time) AS (
    VALUES (9*60*60), (10.5*60*60), (12*60*60), (13.5*60*60), (15*60*60), (16.5*60*60), (18*60*60), (19.5*60*60)
)

SELECT substr(time(available_time, 'unixepoch'), 0, 6)
FROM available_times
WHERE available_time not in (SELECT datetime - $1 FROM records WHERE datetime >= $1 AND datetime <= $1 + 24*60*60)
AND ((available_time > CAST(strftime('%s', 'now') AS INTEGER) - CAST(strftime('%s', date('now')) AS INTEGER)) OR (CAST(strftime('%s', date('now')) AS INTEGER) != $1));`

	rows, err := db.Query(q, result)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	defer rows.Close()

	var datetimes []string
	for rows.Next() {
		var datetime string
		err = rows.Scan(&datetime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		datetimes = append(datetimes, datetime)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}

	return datetimes, nil
}

func UpdateNumber(newNumber, userId int) error {
	q := `UPDATE users SET phone_number=? WHERE user_id=?`

	_, err := db.Exec(q, newNumber, userId)
	if err != nil {
		return fmt.Errorf("failed to update number: %w", err)
	}

	return nil
}

func IsExists(userId int) (bool, error) {
	q := `SELECT COUNT(*) FROM users WHERE user_id=?`

	var count int
	err := db.QueryRow(q, userId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if row exists: %w", err)
	}

	return count > 0, nil
}
