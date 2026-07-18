package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"restapi/internal/models"
	"restapi/pkg/utils"
	"strconv"
	"strings"
)

func GetStudentsDBHandler(students []models.Student, r *http.Request, page, limit int) ([]models.Student, int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, 0, utils.ErrorHandler(err, "Error connectiong to database")
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class FROM students WHERE 1=1"

	var args []interface{}

	//Add pagination
	offset := (page - 1) * limit
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	//for filter
	query, args = utils.AddQueryFilter(r, query, args)

	//sort by fields, possible multiple -> k : sortby , v first_name:asc
	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
		return nil, 0, utils.ErrorHandler(err, "Database query error")
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return nil, 0, utils.ErrorHandler(err, "Error scanning database")
		}
		students = append(students, student)
	}

	//get total all students
	var totalStudents int
	err = db.QueryRow("SELECT COUNT(*) FROM students").Scan(&totalStudents)
	if err != nil {
		totalStudents = 0
		return nil, 0, utils.ErrorHandler(err, "Database query error")
	}
	return students, totalStudents, nil
}

func GetStudentByID(id int) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error connectiong to database")
	}
	defer db.Close()

	var student models.Student
	err = db.QueryRow("SELECT id , first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, "Student not found")
	} else if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Database query error")
	}
	return student, nil
}

func AddStudentsDBHandler(newStudents []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connectiong to database")
	}
	defer db.Close()

	// stmt, err := db.Prepare("INSERT INTO students (first_name, last_name, email, class, subject) VALUES (?, ?, ?, ?, ?)")

	//cara 2
	stmt, err := db.Prepare(utils.GenerateInsertQuery("students", models.Student{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error in prepering query")
	}
	defer stmt.Close()

	addedStudents := make([]models.Student, len(newStudents))
	for i, newStudent := range newStudents {

		//cara 2
		values := utils.GetStructValues(newStudent)
		res, err := stmt.Exec(values...)
		if err != nil {
			if strings.Contains(err.Error(), "a foreign key constraint fails") { //error class not match foreign
				return nil, utils.ErrorHandler(err, "Error constraint class / class teacher does not match")
			} else if strings.Contains(err.Error(), "Duplicate entry ") { //error email unique
				return nil, utils.ErrorHandler(err, err.Error())
			}
			return nil, utils.ErrorHandler(err, "Error inserting data into database")
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error getting last insert ID")
		}
		newStudent.ID = int(lastID)
		addedStudents[i] = newStudent
	}
	return addedStudents, nil
}

func UpdateStudent(id int, reqStudent models.Student) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)

		return models.Student{}, utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.Class)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, utils.ErrorHandler(err, "Student not found")
		}
		return models.Student{}, utils.ErrorHandler(err, "Unable to retrieve data")
	}

	reqStudent.ID = existingStudent.ID
	_, err = db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?", reqStudent.FirstName, reqStudent.LastName, reqStudent.Email, reqStudent.Class, reqStudent.ID)
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error updating data")
	}
	return reqStudent, nil
}

func PatchStudents(updates []map[string]interface{}) error {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	trx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return utils.ErrorHandler(err, "Error starting trx")
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			trx.Rollback()
			return utils.ErrorHandler(err, "Invalid Student ID in update")
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			return utils.ErrorHandler(err, "Unable to convert id to int")
		}

		var studentFromDB models.Student
		err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(
			&studentFromDB.ID,
			&studentFromDB.FirstName,
			&studentFromDB.LastName,
			&studentFromDB.Email,
			&studentFromDB.Class)
		if err != nil {
			log.Println(id)
			log.Println(err)
			trx.Rollback()
			if err == sql.ErrNoRows {
				return utils.ErrorHandler(err, "Student not found")
			}

			return utils.ErrorHandler(err, "Unable to retrieve student")
		}

		//Apply updates using refection
		studentVal := reflect.ValueOf(&studentFromDB).Elem()
		studentType := studentVal.Type()

		for k, v := range update {
			if k == "id" {
				continue //skip updating
			}

			for i := 0; i < studentVal.NumField(); i++ {
				field := studentType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := studentVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(v)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							trx.Rollback()
							log.Printf("cannot convert %v to %v", val.Type(), fieldVal.Type())
							return utils.ErrorHandler(err, "Error updating data")
						}
					}
					break
				}
			}
		}

		_, err = trx.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
			studentFromDB.FirstName,
			studentFromDB.LastName,
			studentFromDB.Email,
			studentFromDB.Class,
			studentFromDB.ID)
		if err != nil {
			trx.Rollback()
			return utils.ErrorHandler(err, "Error updating student")
		}
	}

	//Commit the TRX
	err = trx.Commit()
	if err != nil {
		return utils.ErrorHandler(err, "Error commit the transaction")
	}
	return nil
}

func PatchOneStudent(id int, updates map[string]interface{}) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return models.Student{}, utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(
		&existingStudent.ID,
		&existingStudent.FirstName,
		&existingStudent.LastName,
		&existingStudent.Email,
		&existingStudent.Class)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, utils.ErrorHandler(err, "Student not found")
		}
		return models.Student{

			//Apply update
			// for k, v := range updates {
			// 	switch k {
			// 	case "first_name":
			// 		existingTeacher.FirstName = v.(string)
			// 	case "last_name":
			// 		existingTeacher.LastName = v.(string)
			// 	case "email":
			// 		existingTeacher.Email = v.(string)
			// 	case "class":
			// 		existingTeacher.Class = v.(string)
			// 	case "subject":
			// 		existingTeacher.Subject = v.(string)
			// 	}
		}, nil
	}

	// }

	//Apply updates using reflect
	studentVal := reflect.ValueOf(&existingStudent).Elem()
	studentType := studentVal.Type()

	for k, v := range updates {
		for i := 0; i < studentVal.NumField(); i++ {
			field := studentType.Field(i)
			fmt.Println(field.Tag.Get("json"))
			if field.Tag.Get("json") == k+",omitempty" {
				if studentVal.Field(i).CanSet() {
					fieldVal := studentVal.Field(i)
					fmt.Println("fieldVal:", fieldVal)
					fmt.Println("studentVal.Field(i).Type():", studentVal.Field(i).Type())
					fmt.Println("reflect.ValueOf(v):", reflect.ValueOf(v))
					fieldVal.Set(reflect.ValueOf(v).Convert(studentVal.Field(i).Type()))
				}
			}
		}
	}
	//end using reflect

	_, err = db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
		existingStudent.FirstName,
		existingStudent.LastName,
		existingStudent.Email,
		existingStudent.Class,
		existingStudent.ID)
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error updating data")
	}
	return existingStudent, nil
}

func DeleteOneStudent(id int) error {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "Unable to delete student")
	}

	fmt.Println(res.RowsAffected())
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting student")
	}

	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "Student not found")
	}
	return nil
}

func DeleteStudents(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Error starting transaction")
	}

	stmt, err := tx.Prepare("DELETE FROM students WHERE id = ?")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return nil, utils.ErrorHandler(err, "Error preparing delete statement")
	}
	defer stmt.Close()

	deletedIds := []int{}
	for _, id := range ids {
		res, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			log.Println(err)
			return nil, utils.ErrorHandler(err, "Error deleting student")
		}

		affectedRow, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			log.Println(err)
			return nil, utils.ErrorHandler(err, "Error retrieveing eleted result")
		}

		// if teacher was deleted, the add the ID to deletedIds slices
		if affectedRow > 0 {
			deletedIds = append(deletedIds, id)
		}
		if affectedRow < 1 {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, fmt.Sprintf("ID %d does not exists", id))
		}

	}

	//commit
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Error committing  transaction")
	}

	if len(deletedIds) < 1 {
		return nil, utils.ErrorHandler(err, "IDs do not exist")
	}
	return deletedIds, nil
}
