// @title           ILock HTTP Service API
// @version         1.0
// @description     A comprehensive door access management system with video calling capabilities
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.yourcompany.com/support
// @contact.email  support@yourcompany.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter the token with the `Bearer: ` prefix
package main

import (
	"fmt"
	"ilock-http-service/config"
	"ilock-http-service/models"
	"ilock-http-service/routes"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 初始化日志配置
	if err := config.SetupLogger(); err != nil {
		fmt.Printf("初始化日志配置失败: %v\n", err)
		os.Exit(1)
	}

	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		config.Warning("无法加载.env文件: %v", err)
		// 即使加载失败也继续执行，可能环境变量已经通过其他方式设置
	} else {
		config.Info("成功加载.env文件")
	}

	// 获取配置
	cfg := config.GetConfig()

	// 连接数据库
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}

	// 根据配置执行不同的数据库操作
	if cfg.DBMigrationMode == "drop" {
		// 删除并重建表
		log.Println("警告: 在drop模式下运行，将删除并重建所有表")
		err = dropAndRecreateTables(db)
		if err != nil {
			log.Fatalf("删除并重建表失败: %v", err)
		}
	} else if cfg.DBMigrationMode == "alter" {
		// 执行高级迁移，包括修改列、删除列等
		log.Println("在alter模式下运行，将修改表结构以匹配模型")
		err = advancedMigrate(db, cfg)
		if err != nil {
			log.Fatalf("高级迁移失败: %v", err)
		}
	} else {
		// 默认AutoMigrate，只会添加新列和新表，不会删除或修改列
		log.Println("在标准模式下运行，将只添加新列和新表")
		if err := autoMigrate(db); err != nil {
			log.Fatalf("自动迁移失败: %v", err)
		}
	}

	// 确保系统中有管理员账户
	ensureAdminExists(db, cfg)

	// 初始化路由
	r := routes.SetupRouter(db, cfg)

	// 获取端口配置
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080" // 默认端口
	}

	// 启动服务器
	config.Info("服务器启动在: http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		config.Error("启动服务器失败: %v", err)
		os.Exit(1)
	}
}

// initDB 初始化数据库连接
func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Println("Database connection established")
	return db, nil
}

// autoMigrate 自动迁移所有模型（只添加新列和新表）
func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Admin{},
		&models.PropertyStaff{},
		&models.Device{},
		&models.Resident{},
		&models.CallRecord{},
		&models.AccessLog{},
		&models.EmergencyLog{},
		&models.SystemLog{},
	)

	if err != nil {
		return err
	}

	fmt.Println("Database migration completed")
	return nil
}

// advancedMigrate 执行高级迁移，包括修改列、删除列等
func advancedMigrate(db *gorm.DB, cfg *config.Config) error {
	// 获取底层SQL连接
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get DB connection: %w", err)
	}

	// 禁用外键约束检查
	_, err = sqlDB.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		log.Printf("禁用外键约束检查失败: %v", err)
	}
	defer sqlDB.Exec("SET FOREIGN_KEY_CHECKS = 1") // 确保在函数结束时重新启用外键约束

	// 处理 property_staffs 表的特殊迁移
	log.Println("开始处理property_staffs表的特殊迁移")

	// 1. 检查表是否存在
	var tableExists bool
	err = sqlDB.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = 'property_staffs'", cfg.DBName).Scan(&tableExists)
	if err != nil {
		log.Printf("检查表是否存在失败: %v", err)
	}

	if tableExists {
		// 2. 查询表中的所有列
		rows, err := sqlDB.Query(`
			SELECT COLUMN_NAME, IS_NULLABLE, COLUMN_DEFAULT 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = 'property_staffs'
		`, cfg.DBName)

		if err != nil {
			log.Printf("查询表列失败: %v", err)
		} else {
			defer rows.Close()

			// 定义应该存在于模型中的列名
			modelColumns := map[string]bool{
				"id": true, "phone": true, "property_name": true, "position": true,
				"role": true, "status": true, "remark": true, "username": true,
				"password": true, "created_at": true, "updated_at": true,
				// 不包含 name 和 property_id，它们在模型中已经被移除
			}

			// 处理每一列
			for rows.Next() {
				var columnName, isNullable string
				var columnDefault interface{}
				if err := rows.Scan(&columnName, &isNullable, &columnDefault); err != nil {
					log.Printf("扫描列信息失败: %v", err)
					continue
				}

				// 检查列是否应该在模型中存在
				if !modelColumns[columnName] && columnName != "id" &&
					columnName != "created_at" && columnName != "updated_at" {
					log.Printf("在property_staffs表中发现多余列: %s，准备修改", columnName)

					// 对于property_id列，我们可以先将其设置为可为NULL，再删除
					if columnName == "property_id" && isNullable == "NO" {
						log.Printf("将property_id列修改为可为NULL")
						_, err = sqlDB.Exec("ALTER TABLE property_staffs MODIFY COLUMN property_id INT NULL")
						if err != nil {
							log.Printf("修改property_id列为NULL失败: %v", err)
						}

						// 将所有NULL值更新为0或其他默认值
						_, err = sqlDB.Exec("UPDATE property_staffs SET property_id = 0 WHERE property_id IS NULL")
						if err != nil {
							log.Printf("更新property_id默认值失败: %v", err)
						}
					}

					// 删除多余的列
					log.Printf("正在删除列: %s", columnName)
					_, err = sqlDB.Exec(fmt.Sprintf("ALTER TABLE property_staffs DROP COLUMN %s", columnName))
					if err != nil {
						log.Printf("删除列失败: %v", err)
					} else {
						log.Printf("成功删除列: %s", columnName)
					}
				}
			}
		}
	}

	// 查询所有外键约束
	rows, err := sqlDB.Query(`
		SELECT CONSTRAINT_NAME, TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS 
		WHERE CONSTRAINT_TYPE = 'FOREIGN KEY' 
		AND TABLE_SCHEMA = ?
	`, cfg.DBName)

	if err != nil {
		log.Printf("查询外键约束失败: %v", err)
	} else {
		defer rows.Close()

		// 删除所有找到的外键约束
		for rows.Next() {
			var constraintName, tableName string
			if err := rows.Scan(&constraintName, &tableName); err != nil {
				log.Printf("扫描外键约束信息失败: %v", err)
				continue
			}

			log.Printf("删除外键约束: %s 从表 %s", constraintName, tableName)
			_, err = sqlDB.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP FOREIGN KEY `%s`",
				tableName, constraintName))
			if err != nil {
				log.Printf("删除外键约束失败: %v", err)
			}
		}
	}

	// 执行标准AutoMigrate以添加新列和新表
	return autoMigrate(db)
}

// dropAndRecreateTables 删除并重建所有表
func dropAndRecreateTables(db *gorm.DB) error {
	// 警告: 这将删除所有数据
	log.Println("警告: 正在删除并重建所有表，所有数据将丢失")

	// 禁用外键检查以允许删除表
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	defer db.Exec("SET FOREIGN_KEY_CHECKS = 1")

	// 获取所有表名
	var tables []string
	err := db.Raw("SHOW TABLES").Scan(&tables).Error
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	// 删除所有表
	for _, table := range tables {
		log.Printf("正在删除表: %s", table)
		err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table)).Error
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	// 重新创建所有表
	log.Println("正在重新创建所有表")
	return autoMigrate(db)
}

// ensureAdminExists 确保系统中至少有一个管理员账户
func ensureAdminExists(db *gorm.DB, cfg *config.Config) {
	var count int64
	db.Model(&models.Admin{}).Count(&count)

	// 如果没有管理员，则创建一个默认管理员
	if count == 0 {
		// 生成密码哈希
		defaultPassword := "admin123" // 默认密码
		if cfg.DefaultAdminPassword != "" {
			defaultPassword = cfg.DefaultAdminPassword
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("无法为默认管理员哈希密码: %v", err)
			return
		}

		// 创建默认管理员
		admin := models.Admin{
			Username: "admin",
			Password: string(hashedPassword),
			Email:    "admin@example.com",
			Phone:    "1234567890",
		}

		result := db.Create(&admin)
		if result.Error != nil {
			log.Printf("无法创建默认管理员: %v", result.Error)
			return
		}

		log.Println("已创建默认管理员账户 (用户名: admin)")
	}
}
