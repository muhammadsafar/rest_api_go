package sqlconnect

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"restapi/internal/models"
	"restapi/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-mail/mail/v2"
)

func GetExecsDBHandler(execs []models.Exec, r *http.Request) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connectiong to database")
	}
	defer db.Close()

	query := `SELECT 
		id,
		first_name,
		last_name,
		email,
		username,
		user_created_at,
		inactive_status,
		role 
	FROM execs WHERE 1=1`
	var args []interface{}

	//for filter
	query, args = utils.AddQueryFilter(r, query, args)

	//sort by fields, possible multiple -> k : sortby , v first_name:asc
	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
		return nil, utils.ErrorHandler(err, "Database query error")
	}
	defer rows.Close()

	for rows.Next() {
		var exec models.Exec
		err := rows.Scan(&exec.ID,
			&exec.FirstName,
			&exec.LastName,
			&exec.Email,
			&exec.Username,
			&exec.UserCreatedAt,
			&exec.InactiveStatus,
			&exec.Role)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error scanning database")
		}
		execs = append(execs, exec)
	}
	return execs, nil
}

func GetExecByID(id int) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error connectiong to database")
	}
	defer db.Close()

	row := db.QueryRow(`
		SELECT
			id,
			first_name,
			last_name,
			email,
			username,
			user_created_at,
			inactive_status,
			role
		FROM execs
		WHERE id = ?`, id)

	var exec models.Exec
	err = row.Scan(
		&exec.ID,
		&exec.FirstName,
		&exec.LastName,
		&exec.Email,
		&exec.Username,
		&exec.UserCreatedAt,
		&exec.InactiveStatus,
		&exec.Role,
	)

	if err == sql.ErrNoRows {
		return models.Exec{}, utils.ErrorHandler(err, "Execs not found")
	} else if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Database query error")
	}
	return exec, nil
}

func AddExecsDBHandler(newExecs []models.Exec) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connectiong to database")
	}
	defer db.Close()

	//cara 2
	stmt, err := db.Prepare(utils.GenerateInsertQuery("execs", models.Exec{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error in prepering query")
	}
	defer stmt.Close()

	addedExecs := make([]models.Exec, len(newExecs))
	for i, newExec := range newExecs {

		newExec.Password, err = utils.HashPassword(newExec.Password)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error hashing password")
		}

		//cara 2
		values := utils.GetStructValues(newExec)
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
		newExec.ID = int(lastID)
		addedExecs[i] = newExec
	}
	return addedExecs, nil
}

func PatchExecs(updates []map[string]interface{}) error {
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
			return utils.ErrorHandler(err, "Invalid Execs ID in update")
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			return utils.ErrorHandler(err, "Unable to convert id to int")
		}

		row := db.QueryRow(`
			SELECT
				id,
				first_name,
				last_name,
				email,
				username
			FROM execs
			WHERE id = ?`, id)

		var execFromDB models.Exec
		err = row.Scan(
			&execFromDB.ID,
			&execFromDB.FirstName,
			&execFromDB.LastName,
			&execFromDB.Email,
			&execFromDB.Username,
		)

		if err != nil {
			log.Println(id)
			log.Println(err)
			trx.Rollback()
			if err == sql.ErrNoRows {
				return utils.ErrorHandler(err, "Execs not found")
			}

			return utils.ErrorHandler(err, "Unable to retrieve Execs")
		}

		//Apply updates using refection
		execVal := reflect.ValueOf(&execFromDB).Elem()
		execType := execVal.Type()

		for k, v := range update {
			if k == "id" {
				continue //skip updating
			}

			for i := 0; i < execVal.NumField(); i++ {
				field := execType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := execVal.Field(i)
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

		_, err = trx.Exec(`
			UPDATE execs
			SET
				first_name = ?,
				last_name = ?,
				email = ?,
				username = ?
			WHERE id = ?`,
			execFromDB.FirstName,
			execFromDB.LastName,
			execFromDB.Email,
			execFromDB.Username,
			execFromDB.ID,
		)

		if err != nil {
			trx.Rollback()
			return utils.ErrorHandler(err, "Error updating execs")
		}
	}

	//Commit the TRX
	err = trx.Commit()
	if err != nil {
		return utils.ErrorHandler(err, "Error commit the transaction")
	}
	return nil
}

func PatchOneExec(id int, updates map[string]interface{}) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return models.Exec{}, utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	row := db.QueryRow(`
	SELECT
		id,
		first_name,
		last_name,
		email,
		username
	FROM execs
	WHERE id = ?`, id)

	var existingExec models.Exec
	err = row.Scan(
		&existingExec.ID,
		&existingExec.FirstName,
		&existingExec.LastName,
		&existingExec.Email,
		&existingExec.Username,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Exec{}, utils.ErrorHandler(err, "Exec not found")
		}
		return models.Exec{}, nil
	}

	// }

	//Apply updates using reflect
	execVal := reflect.ValueOf(&existingExec).Elem()
	execType := execVal.Type()

	for k, v := range updates {
		for i := 0; i < execVal.NumField(); i++ {
			field := execType.Field(i)
			fmt.Println(field.Tag.Get("json"))
			if field.Tag.Get("json") == k+",omitempty" {
				if execVal.Field(i).CanSet() {
					fieldVal := execVal.Field(i)
					fmt.Println("fieldVal:", fieldVal)
					fmt.Println("studentVal.Field(i).Type():", execVal.Field(i).Type())
					fmt.Println("reflect.ValueOf(v):", reflect.ValueOf(v))
					fieldVal.Set(reflect.ValueOf(v).Convert(execVal.Field(i).Type()))
				}
			}
		}
	}
	//end using reflect

	_, err = db.Exec(`
			UPDATE execs
			SET
				first_name = ?,
				last_name = ?,
				email = ?,
				username = ?
			WHERE id = ?`,
		existingExec.FirstName,
		existingExec.LastName,
		existingExec.Email,
		existingExec.Username,
		existingExec.ID,
	)
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error updating data")
	}
	return existingExec, nil
}

func DeleteOneExec(id int) error {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM execs WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "Unable to delete execs")
	}

	fmt.Println(res.RowsAffected())
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting execs")
	}

	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "Execs not found")
	}
	return nil
}

func GetUserByUsername(username string) (*models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Unable to connect to database")
	}
	defer db.Close()

	user := &models.Exec{}
	err = db.QueryRow("SELECT id, first_name, last_name, email, username, password, inactive_status, role FROM execs WHERE username = ?", username).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.Password, &user.InactiveStatus, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			// http.Error(w, "User not found", http.StatusNotFound)
			return nil, utils.ErrorHandler(err, "User not found")
		}
		// http.Error(w, "Database query error", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Database query error")
	}
	return user, nil
}

func UpdatePassword(userId int, req models.UpdatePasswordRequest) (string, error) {
	db, err := ConnectDB()
	if err != nil {
		return "", utils.ErrorHandler(err, "database connection error")
	}
	defer db.Close()

	var username, userPassword, userRole string
	err = db.QueryRow("SELECT username, password, role FROM execs WHERE id = ?", userId).Scan(&username, &userPassword, &userRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", utils.ErrorHandler(err, "User not found")
		}
		return "", utils.ErrorHandler(err, "unable to retrieve data")
	}

	err = utils.VerifyPassword(userPassword, req.CurrentPassword)
	if err != nil {
		return "", utils.ErrorHandler(err, "Password you are entered does not match")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return "", utils.ErrorHandler(err, "Error hashing password")
	}

	currentTime := time.Now().Format(time.RFC3339)

	_, err = db.Exec("UPDATE execs SET password = ?, password_change_at = ? WHERE id = ?", hashedPassword, currentTime, userId)
	if err != nil {
		return "", utils.ErrorHandler(err, "failedto update the password")
	}
	return hashedPassword, nil
}

func ForgotPassword(toEmail string) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Internal Error")
	}
	defer db.Close()

	var exec models.Exec
	err = db.QueryRow("SELECT id FROM execs WHERE email = ?", toEmail).Scan(&exec.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorHandler(err, "User not found")
		}
		return utils.ErrorHandler(err, "unable retrieving data")
	}

	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		return utils.ErrorHandler(err, "failed to send password reset email")
	}

	mins := time.Duration(duration)

	expiry := time.Now().Add(mins * time.Minute).Format(time.RFC3339)

	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return utils.ErrorHandler(err, "failed to send password reset email")
	}

	// log.Println("token bytes :", tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	// log.Println("token :", token)

	hashedToken := sha256.Sum256(tokenBytes)
	// log.Println("hashedToken :", hashedToken)

	hashedTokenString := hex.EncodeToString(hashedToken[:])
	// log.Println("hashedTokenString :", hashedTokenString)

	_, err = db.Exec("UPDATE execs SET password_reset_token =?, password_token_expire =? WHERE id =?", hashedTokenString, expiry, exec.ID)
	if err != nil {
		return utils.ErrorHandler(err, "failed to send password reset email")
	}

	//Email the reset email
	resetUrl := fmt.Sprintf("https://localhost:3000/execs/resetpassword/reset/%s", token)
	message := fmt.Sprintf("Forgot your password? Reset with the following link : \n%s\nIf you didn't request a password reset, please ignore this email. This link is only valid for %d minutes.", resetUrl, int(mins))

	m := mail.NewMessage()
	m.SetHeader("From", "admin@school.com") //Replace with your sender email
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Your password reset link")
	m.SetBody("text/plain", message)

	dial := mail.NewDialer("localhost", 1025, "", "")
	err = dial.DialAndSend(m)
	if err != nil {
		return utils.ErrorHandler(err, "failed to send password reset email")
	}
	return nil
}

func ResetPassword(hashedTokenString, password string) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Unable connect to database")
	}
	defer db.Close()

	var user models.Exec

	query := "SELECT id, email FROM execs WHERE password_reset_token = ? AND password_token_expire > ?"
	err = db.QueryRow(query, hashedTokenString, time.Now().Format(time.RFC3339)).Scan(&user.ID, &user.Email)
	if err != nil {
		return utils.ErrorHandler(err, "Invalid retrieve data or expired reset code")
	}

	//hash the new passwword
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return utils.ErrorHandler(err, "Invalid hashed password")
	}

	updateQuery := "UPDATE execs SET password = ?, password_reset_token = NULL, password_token_expire = NULL, password_change_at = ? WHERE id = ?"
	_, err = db.Exec(updateQuery, hashedPassword, time.Now().Format(time.RFC3339), user.ID)
	if err != nil {

		return utils.ErrorHandler(err, "Invalid update user")
	}
	return nil
}
