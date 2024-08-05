package main

import (
	"database/sql"
	"sort"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type UserDevice struct {
	AuthKeyId 	  int    `json:"auth_key_id"`
	DeviceModel   string `json:"device_model"`
	SystemVersion string `json:"system_version"`
	ClientIp 	  string `json:"client_ip"`
	Presence 	  string `json:"updated_at"`
}

func ConnectDB(connectionString string) error {
	database, err := sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}

	if err := database.Ping(); err != nil {
		return err
	}

	db = database
	return nil
}

func GetDB() *sql.DB {
	return db
}

type Bot struct {
	db *sql.DB
}

func (bot *Bot) GetUserByPhoneNumber(phoneNumber string) (int, error) {
	query := "SELECT id FROM users WHERE phone = ?"

	var userId int
	err := bot.db.QueryRow(query, phoneNumber).Scan(&userId)
	if err != nil {
		return -1, err
	}

	return userId, nil
}

func (bot *Bot) GetAuthkeysByUserId(userId int) ([]int, error) {
	query := "SELECT auth_key_id FROM auth_users WHERE user_id = ?"

	rows, err := bot.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authKeyIds []int
	for rows.Next() {
		var authKeyId int
		if err := rows.Scan(&authKeyId); err != nil {
			return nil, err
		}
		authKeyIds = append(authKeyIds, authKeyId)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return authKeyIds, nil
}

func (bot *Bot) GetDevicesByAuthKeyId(authKeys []int) ([]UserDevice, error) {
	query := "SELECT auth_key_id,device_model, system_version, client_ip FROM auths WHERE auth_key_id = ?"

	var devices []UserDevice
	for _, authKey := range authKeys {
		row := bot.db.QueryRow(query, authKey)

		var device UserDevice
		if err := row.Scan(&device.AuthKeyId, &device.DeviceModel, &device.SystemVersion, &device.ClientIp); err != nil {
			return nil, err
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func (bot *Bot) SortAuthkeys(authKeys []int)([]int, error){
	authKeyIds:= make(map[int]time.Time)
	query := `
	SELECT updated_at,perm_auth_key_id
	FROM auth_key_infos 
	WHERE perm_auth_key_id = ?
	ORDER BY id DESC limit 1;
	`	
	for _, authKeyId := range authKeys{
		rows, err := bot.db.Query(query, authKeyId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	
	for rows.Next() {
		var updated_at time.Time
		var authKeyId int
		if err := rows.Scan(&updated_at, &authKeyId); err != nil {
			return nil, err
		}
		authKeyIds[authKeyId] = updated_at
		}
	}

	type kv struct {
		Key   int
		Value time.Time
	}

	var sortedList []kv
	for k, v := range authKeyIds {
		sortedList = append(sortedList, kv{k, v})
	}

	
	sort.Slice(sortedList, func(i, j int) bool {
		return sortedList[i].Value.After(sortedList[j].Value)
	})

	var sortedKeys []int
	for _, kv := range sortedList {
		sortedKeys = append(sortedKeys, kv.Key)
	}

	return sortedKeys, nil
}
