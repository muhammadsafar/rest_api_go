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
)

func GetTeachersDBHandler(teachers []models.Teacher, r *http.Request) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		// http.Error(w, "Error connectiong to database", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error connectiong to database")
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
	var args []interface{}

	//for filter
	query, args = utils.AddQueryFilter(r, query, args)

	//sort by fields, possible multiple -> k : sortby , v first_name:asc
	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
		// http.Error(w, "Database query error", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Database query error")
	}
	defer rows.Close()

	// teachersList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			// http.Error(w, "Error scanning database", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "Error scanning database")
		}
		teachers = append(teachers, teacher)
	}
	return teachers, nil
}

func GetTeacherByID(id int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		// http.Error(w, "Error connectiong to database", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Error connectiong to database")
	}
	defer db.Close()

	var teacher models.Teacher
	err = db.QueryRow("SELECT id , first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
	if err == sql.ErrNoRows {
		// http.Error(w, "Teacher not found", http.StatusNotFound)
		return models.Teacher{}, utils.ErrorHandler(err, "Teacher not found")
	} else if err != nil {
		// http.Error(w, "Database query error", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Database query error")
	}
	return teacher, nil
}

func AddTeachersDBHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		// http.Error(w, "Error connectiong to database", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error connectiong to database")
	}
	defer db.Close()

	// stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?, ?, ?, ?, ?)")

	//cara 2
	stmt, err := db.Prepare(utils.GenerateInsertQuery("teachers", models.Teacher{}))
	if err != nil {
		// http.Error(w, "Error in prepering query", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error in prepering query")
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		//res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)

		//cara 2
		values := utils.GetStructValues(newTeacher)
		res, err := stmt.Exec(values...)
		if err != nil {
			// http.Error(w, "Error inserting data into database", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "Error inserting data into database")
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			// http.Error(w, "Error getting last insert ID", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "Error getting last insert ID")
		}
		newTeacher.ID = int(lastID)
		addedTeachers[i] = newTeacher
	}
	return addedTeachers, nil
}

func UpdateTeacher(id int, reqTeacher models.Teacher) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)

		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err != nil {
		if err == sql.ErrNoRows {
			// http.Error(w, "Teacher no found", http.StatusNotFound)
			return models.Teacher{}, utils.ErrorHandler(err, "Teacher not found")
		}
		// http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Unable to retrieve data")
	}

	reqTeacher.ID = existingTeacher.ID
	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", reqTeacher.FirstName, reqTeacher.LastName, reqTeacher.Email, reqTeacher.Class, reqTeacher.Subject, reqTeacher.ID)
	if err != nil {
		// http.Error(w, "Error updating data", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Error updating data")
	}
	return reqTeacher, nil
}

func PatchTeachers(updates []map[string]interface{}) error {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	trx, err := db.Begin()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Error starting trx", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Error starting trx")
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			trx.Rollback()
			// http.Error(w, "Invalid teacher ID in update", http.StatusBadRequest)
			return utils.ErrorHandler(err, "Invalid teacher ID in update")
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			// http.Error(w, "Unable to convert id to int", http.StatusInternalServerError)
			return utils.ErrorHandler(err, "Unable to convert id to int")
		}

		var teacherFromDB models.Teacher
		err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(
			&teacherFromDB.ID,
			&teacherFromDB.FirstName,
			&teacherFromDB.LastName,
			&teacherFromDB.Email,
			&teacherFromDB.Class,
			&teacherFromDB.Subject)
		if err != nil {
			log.Println(id)
			log.Println(err)
			trx.Rollback()
			if err == sql.ErrNoRows {
				// http.Error(w, "Teacher not found", http.StatusNotFound)
				return utils.ErrorHandler(err, "Teacher not found")
			}

			// http.Error(w, "Unable to retrieve teacher", http.StatusInternalServerError)
			return utils.ErrorHandler(err, "Unable to retrieve teacher")
		}

		//Apply updates using refection
		teacherVal := reflect.ValueOf(&teacherFromDB).Elem()
		teacherType := teacherVal.Type()

		for k, v := range update {
			if k == "id" {
				continue //skip updating
			}

			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := teacherVal.Field(i)
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

		_, err = trx.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
			teacherFromDB.FirstName,
			teacherFromDB.LastName,
			teacherFromDB.Email,
			teacherFromDB.Class,
			teacherFromDB.Subject,
			teacherFromDB.ID)
		if err != nil {
			trx.Rollback()
			// http.Error(w, "Error updating teacher", http.StatusInternalServerError)
			return utils.ErrorHandler(err, "Error updating teacher")
		}
	}

	//Commit the TRX
	err = trx.Commit()
	if err != nil {
		// http.Error(w, "Error commit the transaction", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Error commit the transaction")
	}
	return nil
}

func PatchOneTeacher(id int, updates map[string]interface{}) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(
		&existingTeacher.ID,
		&existingTeacher.FirstName,
		&existingTeacher.LastName,
		&existingTeacher.Email,
		&existingTeacher.Class,
		&existingTeacher.Subject)
	if err != nil {
		if err == sql.ErrNoRows {
			// http.Error(w, "Teacher not found", http.StatusNotFound)
			return models.Teacher{}, utils.ErrorHandler(err, "Teacher not found")
		}
		// http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return models.Teacher{

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
	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
	teacherType := teacherVal.Type()

	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherType.Field(i)
			fmt.Println(field.Tag.Get("json"))
			if field.Tag.Get("json") == k+",omitempty" {
				if teacherVal.Field(i).CanSet() {
					fieldVal := teacherVal.Field(i)
					fmt.Println("fieldVal:", fieldVal)
					fmt.Println("teacherVal.Field(i).Type():", teacherVal.Field(i).Type())
					fmt.Println("reflect.ValueOf(v):", reflect.ValueOf(v))
					fieldVal.Set(reflect.ValueOf(v).Convert(teacherVal.Field(i).Type()))
				}
			}
		}
	}
	//end using reflect

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		existingTeacher.FirstName,
		existingTeacher.LastName,
		existingTeacher.Email,
		existingTeacher.Class,
		existingTeacher.Subject,
		existingTeacher.ID)
	if err != nil {
		// http.Error(w, "Error updating data", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Error updating data")
	}
	return existingTeacher, nil
}

func DeleteOneTeacher(id int) error {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		// http.Error(w, "Unable to delete teacher", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Unable to delete teacher")
	}

	fmt.Println(res.RowsAffected())
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Error deleting teacher")
	}

	if rowsAffected == 0 {
		// http.Error(w, "Teacher no found", http.StatusNotFound)
		return utils.ErrorHandler(err, "Teacher not found")
	}
	return nil
}

func DeleteTeachers(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error starting transaction")
	}

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		// http.Error(w, "Error preparing delete statement", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error preparing delete statement")
	}
	defer stmt.Close()

	deletedIds := []int{}
	for _, id := range ids {
		res, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			log.Println(err)
			// http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "Error deleting teacher")
		}

		affectedRow, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			log.Println(err)
			// http.Error(w, "Error retrieveing eleted result", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "Error retrieveing eleted result")
		}

		// if teacher was deleted, the add the ID to deletedIds slices
		if affectedRow > 0 {
			deletedIds = append(deletedIds, id)
		}
		if affectedRow < 1 {
			tx.Rollback()
			// http.Error(w, fmt.Sprintf("ID %d does not exists", id), http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, fmt.Sprintf("ID %d does not exists", id))
		}

	}

	//commit
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Error committing  transaction", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error committing  transaction")
	}

	if len(deletedIds) < 1 {
		// http.Error(w, "IDs do not exist", http.StatusBadRequest)
		return nil, utils.ErrorHandler(err, "IDs do not exist")
	}
	return deletedIds, nil
}

func GetStudentByTeacherId(teacherId string) (models.Teacher, []models.Student, error) {

	var students []models.Student

	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return models.Teacher{}, nil, utils.ErrorHandler(err, "Error connect to DB")
	}
	defer db.Close()

	var teacher models.Teacher
	err = db.QueryRow("SELECT id , first_name, last_name, email, class, subject FROM teachers WHERE id = ?", teacherId).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
	if err == sql.ErrNoRows {
		log.Println(err)
		//http.Error(w, "Teacher not found", http.StatusNotFound)
		return models.Teacher{}, nil, utils.ErrorHandler(err, "Teacher not found")
	} else if err != nil {
		log.Println(err)
		return models.Teacher{}, nil, utils.ErrorHandler(err, "Error retreiving data teachers")
	}

	query := "SELECT id, first_name, last_name, email, class FROM students WHERE class = (SELECT class FROM teachers WHERE id = ? )"
	rows, err := db.Query(query, teacherId)
	if err != nil {
		log.Println(err)
		return models.Teacher{}, nil, utils.ErrorHandler(err, "Error Retrieving data students")
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			log.Println(err)
			return models.Teacher{}, nil, utils.ErrorHandler(err, "Error scanning data student")
		}
		students = append(students, student)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return models.Teacher{}, nil, utils.ErrorHandler(err, "Error Retrieving data students")
	}
	return teacher, students, nil
}

func TotalStudentByTeacher(teacherId string) (int, error) {
	var studentCount int

	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return 0, utils.ErrorHandler(err, "Unable to connect database")
	}
	defer db.Close()

	query := "SELECT COUNT(*) FROM students WHERE class = (SELECT class FROM teachers WHERE id = ?)"
	err = db.QueryRow(query, teacherId).Scan(&studentCount)
	if err != nil {
		return 0, utils.ErrorHandler(err, "Error to retrieving data")
	}
	return studentCount, nil
}
