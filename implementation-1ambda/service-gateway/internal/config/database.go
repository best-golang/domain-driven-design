package config

import (
	"fmt"

	"github.com/a-trium/domain-driven-design/implementation-1ambda/service-gateway/internal/domain/order"
	"github.com/a-trium/domain-driven-design/implementation-1ambda/service-gateway/internal/domain/product"
	"github.com/a-trium/domain-driven-design/implementation-1ambda/service-gateway/internal/domain/user"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/packr"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rubenv/sql-migrate"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
)

const SqlSchemaDir = "../../asset/sql"

type Database interface {
	Get() *gorm.DB
	EnableDebug()
	Migrate()
	Dialect() string
}

type SQLiteDatabase struct {
	db      *gorm.DB
	dialect string
}

type MySQLDatabase struct {
	db      *gorm.DB
	dialect string
}

func NewSQLiteDatabase(env Environment) Database {
	logger := getDbLogger()

	dialect := "sqlite3"

	uuidString := uuid.NewV4().String()
	filename := fmt.Sprintf("/tmp/ddd_gateway_%s.db", uuidString)
	db, err := gorm.Open(dialect, filename)

	// set gorm options
	db.SingularTable(true)

	if err != nil {
		logger.Fatalw("Failed to connect DB", "error", err)
	}

	return &SQLiteDatabase{db: db, dialect: dialect}
}

func (d *SQLiteDatabase) Get() *gorm.DB {
	return d.db
}

func (d *SQLiteDatabase) EnableDebug() {
	d.db.LogMode(true)
	d.db.Debug()
}

func (d *SQLiteDatabase) Dialect() string {
	return d.dialect
}

func (d *SQLiteDatabase) Migrate() {
	db := d.db
	option := ""
	db.Set("gorm:table_options", option).AutoMigrate(
		&user.User{},
		&user.AuthIdentity{},
		&product.Category{},
		&product.Image{},
		&product.Product{},
		&order.Order{},
		&order.OrderDetail{},
	)

	// SQLite doesn't support `ADD CONSTRAINT`
	// - https://github.com/jinzhu/gorm/blob/b2b568daa8e27966c39c942e5aefc74bcc8af88d/association_test.go#L846
	// db.Model(&user.AuthIdentity{}).AddForeignKey("user_id", "User(id)", "RESTRICT", "CASCADE")

	// Foreign key constraint is disabled by default in SQLite for backward compatibility
	// - http://sqlite.org/foreignkeys.html
	db.Exec("PRAGMA foreign_keys = ON;")
}

func NewMySQLDatabase(env Environment) Database {
	logger := getDbLogger()

	dialect := "mysql"
	dbConnString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		env.MysqlUserName, env.MysqlPassword, env.MysqlHost, env.MysqlPort, env.MysqlDatabase)
	db, err := gorm.Open(dialect, dbConnString)

	if err != nil {
		logger.Fatalw("Failed to connect DB", "error", err)
	}

	// set gorm options
	db.SingularTable(true)

	// set performance related options
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	return &MySQLDatabase{db: db, dialect: dialect}
}

func (d *MySQLDatabase) Dialect() string {
	return d.dialect
}

func (d *MySQLDatabase) Migrate() {
	logger := getDbLogger()

	dialect := "mysql"
	rawDB := d.db.DB()

	migrationSrc := &migrate.PackrMigrationSource{
		Box: packr.NewBox(SqlSchemaDir),
	}

	migrations, err := migrationSrc.FindMigrations()
	if err != nil {
		logger.Fatalw("Failed to find sql migrations", "error", err)
	}

	appliedMigrationCount, err := migrate.Exec(
		rawDB,
		dialect,
		migrate.MemoryMigrationSource{Migrations: migrations},
		migrate.Up,
	)

	if err != nil {
		failedMigr := migrations[appliedMigrationCount]

		logger.Warnw("Found sql migration error. Doing rollback...", "down", failedMigr.Down)
		_, downErr := migrate.Exec(
			rawDB,
			dialect,
			migrate.MemoryMigrationSource{Migrations: []*migrate.Migration{failedMigr}},
			migrate.Down,
		)

		if downErr != nil {
			logger.Errorw("Failed to do rollback sql migration", "error", downErr)
		}

		logger.Fatalw("Failed to do sql migration", "error", err)
	}

	totalMigrationCount := len(migrations)
	if appliedMigrationCount != totalMigrationCount {
		logger.Infow("Some migrations are skipped", "total", totalMigrationCount, "applied", appliedMigrationCount)
	}

	for i := 0; i < totalMigrationCount; i++ {
		skipped := true

		if totalMigrationCount-appliedMigrationCount <= i {
			skipped = false
		}

		logger.Infow("Migration File", "filename", migrations[i].Id, "skip", skipped)
	}

	logger.Infow("Finished migration")
}

func (d *MySQLDatabase) Get() *gorm.DB {
	return d.db
}

func (d *MySQLDatabase) EnableDebug() {
	d.db.LogMode(true)
	d.db.Debug()
}

func GetDatabase() *gorm.DB {
	logger := getDbLogger()

	env := Env
	useSqlite := env.IsTestMode()

	var database Database

	// Use sqlite3 for `TEST` env
	if useSqlite {
		database = NewSQLiteDatabase(env)
	} else {
		database = NewMySQLDatabase(env)
	}

	logger.Infow("Database connected", "dialect", database.Dialect())

	if (env.IsLocalMode() && env.DebugSQLEnabled()) || (env.IsTestMode() && env.DebugSQLEnabled()) {
		database.EnableDebug()
	}

	// migration
	database.Migrate()

	return database.Get()
}

func getDbLogger() *zap.SugaredLogger {
	logger := GetLogger().With("context", "database")

	return logger
}
