package pkg

import (
	"encoding/json"
	"errors"
	"os"
)

// LoadEmployees загружает список сотрудников из JSON-файла.
func LoadEmployees(filePath string) (*[]Employee, error) {
	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// создать файл
		_, err := os.Create(filePath)
		if err != nil {
			return nil, errors.New("не удалось создать файл: " + err.Error())
		}
	}

	// Чтение содержимого файла
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.New("не удалось прочитать файл: " + err.Error())
	}

	// Декодирование содержимого файла в срез структур Employee
	var employees []Employee

	// если файл пустой - вернуть пустой срез
	if len(fileContent) == 0 {
		return &employees, nil
	}

	if err := json.Unmarshal(fileContent, &employees); err != nil {
		return nil, errors.New("не удалось декодировать JSON: " + err.Error())
	}

	return &employees, nil
}

// SaveEmployees сохраняет обновленный список сотрудников в JSON-файл.
func SaveEmployees(filePath string, employees *[]Employee) error {
	// Кодирование среза структур Employee в формат JSON
	jsonData, err := json.MarshalIndent(&employees, "", "    ")
	if err != nil {
		return errors.New("не удалось закодировать в JSON: " + err.Error())
	}

	// Запись закодированных данных в файл
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return errors.New("не удалось сохранить данные в файл: " + err.Error())
	}

	return nil
}

// LoadDutyHistory загружает исторические данные из файла.
func LoadDutyHistory(filePath string) (*DutyHistoryStorage, error) {
	var storage DutyHistoryStorage

	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// создать файл
		_, err := os.Create(filePath)
		if err != nil {
			return &storage, errors.New("не удалось создать файл: " + err.Error())
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return &storage, err
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, &storage)
		if err != nil {
			return &storage, err
		}
	}

	return &storage, nil
}

// SaveDutyHistory сохраняет исторические данные в файл.
func SaveDutyHistory(filePath string, storage *DutyHistoryStorage) error {
	data, err := json.MarshalIndent(&storage, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Сохраняет cсгенерированное расписание в файл
func SaveShedule(filePath string, storage *DutyHistory) error {
	data, err := json.MarshalIndent(&storage, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
