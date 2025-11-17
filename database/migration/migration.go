package migration

import (
	"clean-arch/database"
	"clean-arch/pkg/util"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
)

type migrations struct {
	gorm.Model
	Migrations string `json:"migrations"`
}

func FirstMigrate() {

	conn := database.GetConnection()

	if err := CopyBaseMigrations(); err != nil {
		log.Println("error copying base migration:", err)
	}

	conn.AutoMigrate(&migrations{})
}

func CreateMigrationFile(migrationFileName string) error {
	var timeNow = time.Now()
	var year int = timeNow.Year()
	var month int = int(timeNow.Month())
	var day int = timeNow.Day()
	var hour, minute, seconds int = timeNow.Clock()
	var format = fmt.Sprintf("%d-%d-%d_%d:%d:%d_%s.sql", year, month, day, hour, minute, seconds, migrationFileName)

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	migrationsPath := wd + "/database/migration/migrations_file/"
	file, err := os.Create(migrationsPath + format)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

func Migrate(migrationFileName string) error {
	wd, files, err := findMigrationFiles()
	if err != nil {
		return err
	}

	if len(files) < 1 {
		return fmt.Errorf("migration file not found")
	}

	var isFileExist bool
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == migrationFileName {
			isFileExist = true
		}
	}

	if !isFileExist {
		return fmt.Errorf("migration file not found")
	}

	migrateTable := []migrations{}
	conn := database.GetConnection() // Get db connection
	err = conn.Raw("SELECT migrations FROM migrations").Find(&migrateTable).Error
	if err != nil {
		return err
	}

	for _, val := range migrateTable {
		if migrationFileName == val.Migrations {
			return fmt.Errorf("nothing to migrate")
		}
	}

	migrationLog(1, migrationFileName)

	err = executeSQLQueryFile(wd, migrationFileName, conn)
	if err != nil {
		migrationLog(3, migrationFileName)
		return err
	}

	migrationLog(2, migrationFileName)

	return nil
}

func MigrateAll() error {
	wd, files, err := findMigrationFiles()
	if err != nil {
		return err
	}

	if len(files) < 1 {
		return fmt.Errorf("migration file not found")
	}

	migrateTable := []migrations{}
	conn := database.GetConnection() // Get db connection
	err = conn.Raw("SELECT migrations FROM migrations").Find(&migrateTable).Error
	if err != nil {
		return err
	}

	var migrationsDataMap = map[string]bool{}
	for _, mt := range migrateTable {
		migrationsDataMap[mt.Migrations] = true
	}

	for _, val := range files {
		checkExist := migrationsDataMap[val.Name()]

		if checkExist {
			continue
		}

		migrationLog(1, val.Name())

		err = executeSQLQueryFile(wd, val.Name(), conn)
		if err != nil {
			migrationLog(3, val.Name())
			return err
		}

		migrationLog(2, val.Name())
	}
	return nil
}

func findMigrationFiles() (string, []fs.DirEntry, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}

	migrationsPath := wd + "/database/migration/migrations_file/"
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		defer os.Exit(1)
		return "", nil, err
	}
	return wd, files, nil
}

func executeSQLQueryFile(wd string, migrationFileName string, conn *gorm.DB) error {
	mFile, err := os.Open(wd + "/database/migration/migrations_file/" + migrationFileName)
	if err != nil {
		defer os.Exit(1)
		return err
	}

	defer mFile.Close()

	stat, err := mFile.Stat()
	if err != nil {
		defer os.Exit(1)
		return err
	}

	buffer := make([]byte, stat.Size())

	_, err = mFile.Read(buffer)
	if err != nil {
		defer os.Exit(1)
		return err
	}

	tx := conn.Begin()
	err = tx.Exec(string(buffer)).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	newMigrationHistory := migrations{
		Migrations: migrationFileName,
	}

	err = tx.Model(&migrations{}).Create(&newMigrationHistory).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func migrationLog(state int, fileName string) {
	colorBlue := "\033[34m"
	colorGreen := "\033[32m"
	colorRed := "\033[31m"

	switch state {
	case 1:
		log.Println(colorBlue, "Migrating:", fileName)
		return
	case 2:
		log.Println(colorGreen, "success")
		return
	case 3:
		log.Println(colorRed, "Failed to migrate:", fileName)
		return
	}
}

func CopyBaseMigrations() error {
	// detect db driver
	dbDriver := util.GetEnv("DB_DRIVER", "mysql")

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	var basePath string
	switch dbDriver {
	case "psql":
		basePath = wd + "/database/migration/migration_base/psql/"
	default:
		basePath = wd + "/database/migration/migration_base/mysql/"
	}

	destPath := wd + "/database/migration/migrations_file/"

	files, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed read base migration folder: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		srcFile := basePath + file.Name()
		destFile := destPath + file.Name()

		// Skip if exists
		if _, err := os.Stat(destFile); err == nil {
			log.Println("\033[33m", "Skip (exists):", file.Name())
			continue
		}

		err := copyFile(srcFile, destFile)
		if err != nil {
			return fmt.Errorf("copy error %s: %w", file.Name(), err)
		}

		log.Println("\033[32m", "Copied:", file.Name())
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
