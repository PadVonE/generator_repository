package service

import (
	"encoding/json"
	"fmt"
	"generator/entity"
	"generator/generators/repository"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type MigrationResponse struct {
	Code     string
	Filename string
}

func (s *Service) GenerateMigrationApi(ctx *gin.Context) {
	log.Println("\033[35m", "\n\nMigration file", "\033[0m")

	projectID := ctx.Query("project_id")

	project := entity.Project{}

	err := s.DB.First(&project, "id = ?", projectID).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	servicePath := filepath.FromSlash(project.LocalPath)

	projectComponents := entity.ProjectComponents{}
	err = json.Unmarshal([]byte(project.LastStructure), &projectComponents)

	migration := ""
	//Таблица для отслеживания изменений
	migrationEditLog := "create table if not exists edited_log(\n" +
		"    id serial not null constraint edited_log_event_pkey primary key,\n" +
		"    created_at timestamp not null default CURRENT_TIMESTAMP,\n" +
		"    action text not null,\n" +
		"    table_name text not null,\n" +
		"    table_id integer not null,\n" +
		"    edited_user_id integer not null,\n" +
		"    json_string json not null\n);\n\n"

	// Инлекс
	migrationEditLog += "create index if not exists edited_log_table_name_table_id_idx on edited_log (table_name, table_id);\n\n"

	// Тригер
	migrationEditLog += "create or replace function edited_user_id() returns trigger\n" +
		"    language plpgsql\nas\n$$\nbegin\n" +
		"    if new.edited_user_id > 0 then\n" +
		"        insert into \"edited_log\" (\"action\", \"table_name\", \"table_id\", \"edited_user_id\", \"json_string\")\n" +
		"        values (tg_op, tg_table_name, new.id, new.edited_user_id, row_to_json(new.*));\n" +
		"    end if;\n" +
		"    return new;\nend;\n$$;\n\n"

	hasEditedLog := false
	for _, l := range projectComponents.ListOfStruct {
		if l.Type == entity.TypeMain {
			// Если в entity нет поля id то пропускаем
			hasId := false
			for _, row := range l.Rows {
				if row.Name == "Id" {
					hasId = true
				}
			}
			if hasId == false {
				continue
			}

			code, addEditLogTrigger, err := generators.GenerateMigration(l)
			if err != nil {
				log.Error(err)
				continue
			}

			if addEditLogTrigger == true {
				hasEditedLog = true
			}
			migration += code
		}
	}
	now := time.Now()

	if hasEditedLog {
		migration = migrationEditLog + migration
	}

	saveFilePath := servicePath + "/migrations/"
	saveFileName := strconv.Itoa(int(now.Unix())) + "_init.up.sql"

	// Проверим есть ли файл миграции
	migrationIsset := false

	entries, err := os.ReadDir(saveFilePath)
	if err == nil {

		for _, e := range entries {
			log.Println(e.Name())
			contain := strings.Contains(e.Name(), "_init")
			if contain {
				migrationIsset = true
				saveFileName = e.Name()
			}
		}
	}

	if migrationIsset {
		log.Warn("Migration isset in path " + saveFilePath + saveFileName)
	}

	formattedCodeOldCode := ""
	hasFile := false
	hasDiff := true
	if _, err := os.Stat(saveFilePath + saveFileName); err == nil {
		hasFile = true

		file, err := os.ReadFile(saveFilePath + saveFileName)
		if err != nil {
			log.Fatalf("Ошибка при чтении файла: %v", err)
		}

		formattedCodeOldCode = string(file)

		if CompareStrings(formattedCodeOldCode, migration) {
			hasDiff = false
		}
	}

	response := []entity.FilesPreview{{
		FilePath: saveFilePath + saveFileName,
		NewCode:  migration,
		OldCode:  formattedCodeOldCode,
		HasFile:  hasFile,
		HasDiff:  hasDiff,
	}}

	ctx.JSON(200, response)

}
