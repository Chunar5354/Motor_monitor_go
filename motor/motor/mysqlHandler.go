package motor

import (
	"database/sql"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// the length of every parameters while createing tables
var valueLength = map[string]string{
	"current_A":           "1800",
	"current_B":           "1800",
	"current_C":           "1800",
	"voltage_AB":          "1800",
	"voltage_BC":          "1800",
	"voltage_CA":          "1800",
	"temp_fore_winding_A": "40",
	"temp_fore_winding_B": "40",
	"temp_fore_winding_C": "40",
	"temp_rotator":        "40",
	"temp_water":          "40",
	"temp_fore_bearing":   "40",
	"temp_rear_bearing":   "40",
	"temp_controller_env": "40",
	"vib_fore_bearing":    "8000",
	"vib_rear_bearing":    "8000",
	"rev":                 "800",
}

type MysqlInfo struct {
	user     string
	password string
	port     string
}

var mysqlInfo MysqlInfo

// set mysql information by user
func MysqlInit(user, password, port string) {
	mysqlInfo.user = user
	mysqlInfo.password = password
	mysqlInfo.port = port
}

// set a time intercal to determine if need to create new table in mysql
var lastDay = timeInit()
var OneDay = int64(86400) // 86400 secoonds in one day

func timeInit() time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", "2020-12-27 00:00:00")
	return t
}

// keep a value: lastTime to tell the fetch server which is the newest value
func updateLastTime(db *sql.DB, lastTime string) {
	sqlString := "update `info` set value = '" + lastTime + "' where field = 'lastTime'"
	res, err := db.Exec(sqlString)
	if err != nil {
		// log.Println("Update failed: ", err)
		Error.Println("Update failed: ", err)
		return
	}
	countAff, err := res.RowsAffected()
	if err != nil {
		// log.Println("RowsAffected failed:", err)
		Error.Println("RowsAffected failed:", err)
		return
	}
	if countAff == 0 { // if lastTime doesn't exist, insert it
		sqlString = "insert into `info` values ('lastTime', '" + lastTime + "')"
		_, err := db.Exec(sqlString)
		if err != nil {
			// log.Println("Insert failed: ", err)
			Error.Println("Insert failed: ", err)
			return
		}
	}
}

// create today's table
func createTable(db *sql.DB, para, ymd string) {
	sqlString := "create table `" + para + "_" + ymd + "` ( `create_time` TIME NOT NULL, `value` varchar(" + valueLength[para] + ") NOT NULL, PRIMARY KEY (`create_time`) ) ENGINE=InnoDB;"
	if _, err := db.Exec(sqlString); err != nil {
		// log.Println("create table failed:", err)
		Error.Println("create table failed:", err)
		return
	}
	// fmt.Println("create table successd: ", para)
}

// check if today's table is existed
func tableExist(db *sql.DB, tableName string) bool {
	_, err := db.Query("select value from `" + tableName + "` limit 1") // ? = placeholder
	if err != nil {
		// log.Println(tableName, "doesn't exist: ", err) // proper error handling instead of panic in your app
		Error.Println(tableName, "doesn't exist: ", err) // proper error handling instead of panic in your app
		return false
	} else {
		// fmt.Printf("%T, %v\n", res, res)
		return true
	}
}

// insert data into mysql
func mysqlInsert(db *sql.DB, n *sync.WaitGroup, para, ymd, hms, data string) {
	defer n.Done()
	tableName := para + "_" + ymd
	// if the upload time is not today, need to check if the tables are existed
	if time.Now().Unix()-lastDay.Unix() > OneDay {
		if !tableExist(db, tableName) {
			createTable(db, para, ymd)
		}
		lastDay, _ = time.Parse("2006-01-02", ymd) // modify lastDay to today
	}
	sqlString := "insert into `" + tableName + "` values(?, ?)"
	_, err := db.Exec(sqlString, hms, data[:len(data)-1])
	if err != nil {
		// log.Printf("Insert data failed in %v, err:%v\n", para, err)
		Error.Printf("Insert data failed in %v, err:%v\n", para, err)
		// if fail to insert, try create table
		createTable(db, para, ymd)
		return
	}
}

// get the last time when upload data, then use lastTime to fetch data
func getLastTime(db *sql.DB) string {
	sqlString := "select value from info where field = 'lastTime'"
	var lastTime string
	res := db.QueryRow(sqlString)
	if err := res.Scan(&lastTime); err != nil {
		// log.Println("Get last time failed: ", err)
		Error.Println("Get last time failed: ", err)
	}
	return lastTime

}

// fetch data from the table and write data into response
func mysqlFetch(db *sql.DB, para, ymd, hms string) []string {
	var data string // receive data from table
	tableName := para + "_" + ymd
	sqlString := "select value from `" + tableName + "` where create_time = '" + hms + "'"
	res := db.QueryRow(sqlString)
	if err := res.Scan(&data); err != nil {
		// log.Println("Failed to fetch data from mysql: ", err)
		Error.Println("Failed to fetch data from mysql: ", err)
	}
	return strings.Split(data, "/") // return an array of data
}

func mysqlUpload(req map[string]string, serial_number, create_time string) error {
	// set for mysql connection
	connection := mysqlInfo.user + ":" + mysqlInfo.password + "@tcp(localhost:" + mysqlInfo.port + ")/motor_" + serial_number
	db, err := sql.Open("mysql", connection)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	ymd := create_time[:10] // year-month-day, use in table name in mysql
	hms := create_time[11:] // hour-minute-second, use as primary key in mysql

	go updateLastTime(db, create_time)

	var n sync.WaitGroup
	for para, data := range req {
		n.Add(1)
		go mysqlInsert(db, &n, para, ymd, hms, data)
		n.Add(1)
		go faultUpload(db, &n, para, data[:11], create_time)
	}
	n.Wait()
	return nil
}

// connect to mysql and cak fetch data function with the request parameters
func handleFetchSql(serialNumber, startTime, onlyFault string, parameters []string) string {
	connection := mysqlInfo.user + ":" + mysqlInfo.password + "@tcp(localhost:" + mysqlInfo.port + ")/motor_" + serialNumber
	db, err := sql.Open("mysql", connection)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// set for mysql connection
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// today := time.Now().Format("2006-01-02") // use today to construct table name
	m := make(map[string]interface{})
	createTime := startTime
	if createTime == "" {
		createTime = getLastTime(db)
	}

	// 只查询故障数据
	if onlyFault == "true" {
		return fetchOnlyFault(db, createTime, parameters)
	}

	m["create_time"] = createTime
	ymd := createTime[:10] // year-month-day, use in table name in mysql
	hms := createTime[11:] // hour-minute-second, use as primary key in mysql
	faultMsg := make(map[string]string)
	m["status"] = 200  // ok status
	for _, para := range parameters {
		data := mysqlFetch(db, para, ymd, hms)
		m[para] = data
		fault, faultValue := faultFetch(para, data[0])  // add fault values
		if fault {
			m["status"] = 400  // fault status
			faultMsg[para] = faultValue
		}
	}
	m["fault"] = faultMsg
	response, err := json.Marshal(m)
	if err != nil {
		// log.Fatalf("JSON marshaling failed: %s", err)
		Error.Fatalf("JSON marshaling failed: %s", err)
	}
	return string(response)
}
