package motor

import (
	"database/sql"
	"strconv"
	"strings"
	"sync"
)

var maxThreshold = map[string]float64{
	"current_A":           200,
	"current_B":           200,
	"current_C":           200,
	"voltage_AB":          3500,
	"voltage_BC":          3500,
	"voltage_CA":          3500,
	"temp_fore_winding_A": 80,
	"temp_fore_winding_B": 80,
	"temp_fore_winding_C": 80,
	"temp_rotator":        80,
	"temp_water":          80,
	"temp_fore_bearing":   80,
	"temp_rear_bearing":   80,
	"temp_controller_env": 50,
	"vib_fore_bearing":    100,
	"vib_rear_bearing":    100,
	"rev":                 3000,
}

var minThreshold = map[string]float64{
	"current_A":           30,
	"current_B":           30,
	"current_C":           30,
	"voltage_AB":          1000,
	"voltage_BC":          1000,
	"voltage_CA":          1000,
	"temp_fore_winding_A": 0,
	"temp_fore_winding_B": 0,
	"temp_fore_winding_C": 0,
	"temp_rotator":        0,
	"temp_water":          0,
	"temp_fore_bearing":   0,
	"temp_rear_bearing":   0,
	"temp_controller_env": 0,
	"vib_fore_bearing":    0,
	"vib_rear_bearing":    0,
	"rev":                 100,
}

// check if ther is fault data while uploading data, if there is, insert faule data into fault table
func faultUpload(db *sql.DB, n *sync.WaitGroup, para, data, create_time string) {
	defer n.Done()
	valueString := strings.Split(data, "/")
	value, err := strconv.ParseFloat(valueString[0], 64)
	if err != nil {
		// log.Println("Failed to converse value: ", para, err)
		Error.Println("Failed to converse value: ", para, err)
		return
	}
	if value < minThreshold[para] || value > maxThreshold[para] {
		// insert falue value into mysql
		sqlString := "insert into fault values (?, ?)"
		_, err := db.Exec(sqlString, "'"+para+create_time+"'", valueString[0])
		if err != nil {
			//log.Println("Failed to insert falut: ", para, err)
			Error.Println("Failed to insert falut: ", para, err)
			return
		}
		// do something, such as sending error data, to be continued...
	}
}

// while fetching data, if there is fault, send it back to client
func faultFetch(para, data string) (bool, []string) {
	value, err := strconv.ParseFloat(data, 64)
	if err != nil {
		//log.Println("Failed to converse value: ", para, err)
		Error.Println("Failed to converse value: ", para, err)
		return false, nil
	}
	if value < minThreshold[para] || value > maxThreshold[para] {
		return true, []string{para, data}
	}
	return false, nil
}
