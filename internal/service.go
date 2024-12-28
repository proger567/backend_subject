package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"testgenerate_backend_subject/internal/app"
	"time"
)

type Service interface {
	GetSubjects(ctx context.Context) ([]app.Subject, error)
	AddSubject(ctx context.Context, subjectAdd app.Subject) error
	UpdateSubject(ctx context.Context, subject app.Subject) error
	DeleteSubject(ctx context.Context, id int) error
}

type userService struct {
	logger *logrus.Logger
}

func NewBasicService(logger *logrus.Logger) Service {
	return userService{
		logger: logger,
	}
}

func NewService(logger *logrus.Logger, requestCount metrics.Counter, requestLatency metrics.Histogram) Service {
	var svc Service
	{
		svc = NewBasicService(logger)
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(requestCount, requestLatency)(svc)
	}
	return svc
}

// ----------------------------------------------------------------------------------------------------------------------
func (u userService) GetSubjects(ctx context.Context) ([]app.Subject, error) {
	var subjects []app.Subject
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s"+
		" password=%s dbname=%s sslmode=disable",
		app.GetEnv("DB_HOST", "localhost"), app.GetEnvAsInt("DB_PORT", 5432),
		app.GetEnv("DB_USER", "postgres"), app.GetEnv("DB_PASSWORD", "pgpassword"),
		app.GetEnv("DB_NAME", "generate"))
	conn, err := pgx.Connect(ctx, psqlInfo)
	if err != nil {
		erRet := fmt.Errorf("GetSubjects. Unable to connect to database: %v\n", err)
		return subjects, erRet
	}
	defer conn.Close(ctx)

	rows, errRows := conn.Query(ctx, `select to_json(t.*)
					from (select id,comment,date_create, description , last_time_update, name, type, parent_id from subject) t`)
	if errRows != nil {
		erResp := fmt.Errorf("GetSubjects QueryRow: %v\n", errRows)
		return subjects, erResp
	}

	for rows.Next() {
		var res string
		errScan := rows.Scan(&res)
		if errScan != nil {
			erRet := fmt.Errorf("GetSubjects rows.Scan: %v\n", errScan)
			return subjects, erRet
		}
		var result app.Subject
		errU := json.Unmarshal([]byte(res), &result)
		if errU != nil {
			erRet := fmt.Errorf("GetSubjects json.Unmarshal: %v\n", errU)
			return subjects, erRet
		}
		subjects = append(subjects, result)
	}
	return subjects, nil
}

func (u userService) AddSubject(ctx context.Context, subject app.Subject) error {
	var errA error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s"+
		" password=%s dbname=%s sslmode=disable",
		app.GetEnv("DB_HOST", "localhost"), app.GetEnvAsInt("DB_PORT", 5432),
		app.GetEnv("DB_USER", "postgres"), app.GetEnv("DB_PASSWORD", "pgpassword"),
		app.GetEnv("DB_NAME", "generate"))
	conn, err := pgx.Connect(ctx, psqlInfo)
	if err != nil {
		erRet := fmt.Errorf("AddSubject. Unable to connect to database: %v\n", err)
		return erRet
	}
	defer conn.Close(ctx)

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		errA = fmt.Errorf("AddSubject conn.BeginTx %v\n", err)
		return errA
	}
	defer func() {
		if errA != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	//All users add with role == 'user'
	//Next Administrator may change this role
	//SuperAdmins insert trough database
	_, err = tx.Exec(ctx, `insert into subject (comment,date_create, description , last_time_update, name, type, parent_id) values($1, $2, $3, $4, $5, $6, $7)`,
		subject.Comment, time.Now(), subject.Description, subject.Description, time.Now(), subject.Name, subject.Type, subject.ParentID)
	if err != nil {
		errA = fmt.Errorf("AddSubject insert into subject: %v\n", err)
		return errA
	}
	return nil
}
func (u userService) UpdateSubject(ctx context.Context, subject app.Subject) error {
	var errU error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s"+
		" password=%s dbname=%s sslmode=disable",
		app.GetEnv("DB_HOST", "localhost"), app.GetEnvAsInt("DB_PORT", 5432),
		app.GetEnv("DB_USER", "postgres"), app.GetEnv("DB_PASSWORD", "pgpassword"),
		app.GetEnv("DB_NAME", "generate"))
	conn, err := pgx.Connect(ctx, psqlInfo)
	if err != nil {
		errU = fmt.Errorf("UpdateSubject. Unable to connect to database: %v\n", err)
		return errU
	}
	defer conn.Close(ctx)

	_, errU = conn.Exec(ctx, `update subject set comment = $2, description = $3 , last_time_update = $4, name = $5, type = $6, parent_id = $7 where id = $1`,
		subject.ID, subject.Comment, subject.Description, time.Now(), subject.Name, subject.Type, subject.ParentID)
	if errU != nil {
		return fmt.Errorf("UpdateSubject conn.Exec: %v\n", errU)
	}
	return nil
}
func (u userService) DeleteSubject(ctx context.Context, id int) error {
	var errD error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s"+
		" password=%s dbname=%s sslmode=disable",
		app.GetEnv("DB_HOST", "localhost"), app.GetEnvAsInt("DB_PORT", 5432),
		app.GetEnv("DB_USER", "postgres"), app.GetEnv("DB_PASSWORD", "pgpassword"),
		app.GetEnv("DB_NAME", "generate"))
	conn, err := pgx.Connect(ctx, psqlInfo)
	if err != nil {
		errD = fmt.Errorf("DeleteSubject. Unable to connect to database: %v\n", err)
		return errD
	}
	defer conn.Close(ctx)

	_, errD = conn.Exec(ctx, `delete from subject where id = $1`, id)
	if errD != nil {
		return fmt.Errorf("DeleteSubject conn.Exec: %v\n", errD)
	}

	return nil
}
