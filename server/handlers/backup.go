package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"dootask-kpi-server/global"
	"dootask-kpi-server/models"
	"dootask-kpi-server/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 备份目录
const BackupDir = "/web/db/backups"

// 备份响应结构
type BackupResponse struct {
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

// 备份历史响应结构
type BackupHistoryResponse struct {
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	CreatedAt   time.Time `json:"created_at"`
	FileType    string    `json:"file_type"`    // "db" 或 "sql"
	DownloadURL string    `json:"download_url"` // 下载URL
}

// 创建数据库备份 (SQL格式)
func CreateBackup(c *gin.Context) {
	// 创建备份目录
	if err := os.MkdirAll(BackupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建备份目录失败",
		})
		return
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("backup_%s.sql", timestamp)
	backupPath := filepath.Join(BackupDir, fileName)

	// 创建备份文件
	backupFile, err := os.Create(backupPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建备份文件失败",
		})
		return
	}
	defer backupFile.Close()

	// 生成SQL备份内容
	if err := generateSQLBackup(backupFile); err != nil {
		// 清理失败的备份文件
		os.Remove(backupPath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成SQL备份失败: " + err.Error(),
		})
		return
	}

	// 获取文件信息
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		os.Remove(backupPath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取备份文件信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, BackupResponse{
		FileName:  fileName,
		FileSize:  fileInfo.Size(),
		CreatedAt: fileInfo.ModTime(),
		Message:   "备份创建成功",
	})
}

// 获取备份历史列表
func GetBackupHistory(c *gin.Context) {
	// 检查备份目录是否存在
	if _, err := os.Stat(BackupDir); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{
			"data":  []BackupHistoryResponse{},
			"total": 0,
		})
		return
	}

	// 读取备份目录
	files, err := os.ReadDir(BackupDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "读取备份目录失败",
		})
		return
	}

	// 过滤备份文件并获取信息（支持.db和.sql格式）
	var backups []BackupHistoryResponse
	for _, file := range files {
		fileName := file.Name()
		if !file.IsDir() && strings.HasPrefix(fileName, "backup_") &&
			(strings.HasSuffix(fileName, ".db") || strings.HasSuffix(fileName, ".sql")) {
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}

			// 确定文件类型
			fileType := "db"
			if strings.HasSuffix(fileName, ".sql") {
				fileType = "sql"
			}

			downloadURL := utils.GetFileURL(c.GetString("base_url"), fmt.Sprintf("/api/download/backups/%s", fileName))

			backups = append(backups, BackupHistoryResponse{
				FileName:    fileName,
				FileSize:    fileInfo.Size(),
				CreatedAt:   fileInfo.ModTime(),
				FileType:    fileType,
				DownloadURL: downloadURL,
			})
		}
	}

	// 按创建时间倒序排列，最新的在前面
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	c.JSON(http.StatusOK, gin.H{
		"data":  backups,
		"total": len(backups),
	})
}

// 生成备份下载URL
func GenerateBackupDownloadURL(c *gin.Context) {
	fileName := c.Param("filename")

	// 验证文件名格式（支持.db和.sql文件）
	if !strings.HasPrefix(fileName, "backup_") ||
		(!strings.HasSuffix(fileName, ".db") && !strings.HasSuffix(fileName, ".sql")) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的备份文件名",
		})
		return
	}

	// 检查备份文件是否存在
	filePath := filepath.Join(BackupDir, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "备份文件不存在",
		})
		return
	}

	// 生成随机key并缓存
	randomKey := "backup_" + uuid.New().String()
	global.Cache.Set(randomKey, fileName, time.Minute*5)

	// 返回下载URL
	downloadURL := utils.GetFileURL(c.GetString("base_url"), fmt.Sprintf("/api/download/backups/%s", randomKey))

	c.JSON(http.StatusOK, gin.H{
		"download_url": downloadURL,
		"file_name":    fileName,
	})
}

// 下载备份文件（参考导出功能的下载机制）
func DownloadBackup(c *gin.Context) {
	fileName := c.Param("randomKey")

	filePath := filepath.Join(BackupDir, fileName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "备份文件不存在",
		})
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)

	// 根据文件扩展名设置Content-Type
	if strings.HasSuffix(fileName, ".sql") {
		c.Header("Content-Type", "application/sql")
	} else {
		c.Header("Content-Type", "application/x-sqlite3")
	}

	// 返回文件
	c.File(filePath)
}

// 恢复数据库
func RestoreBackup(c *gin.Context) {
	fileName := c.Param("filename")

	// 验证文件名格式，支持.db和.sql文件恢复
	if !strings.HasPrefix(fileName, "backup_") ||
		(!strings.HasSuffix(fileName, ".db") && !strings.HasSuffix(fileName, ".sql")) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的备份文件名",
		})
		return
	}

	backupPath := filepath.Join(BackupDir, fileName)

	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "备份文件不存在",
		})
		return
	}

	// 根据文件类型执行不同的恢复逻辑
	var err error
	if strings.HasSuffix(fileName, ".db") {
		// .db文件：直接替换数据库文件
		err = restoreFromDBFile(backupPath)
	} else if strings.HasSuffix(fileName, ".sql") {
		// .sql文件：执行SQL脚本
		err = restoreFromSQLFile(backupPath)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "恢复数据库失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据库恢复成功",
	})
}

// 删除备份文件
func DeleteBackup(c *gin.Context) {
	fileName := c.Param("filename")

	// 验证文件名格式（支持.db和.sql文件）
	if !strings.HasPrefix(fileName, "backup_") ||
		(!strings.HasSuffix(fileName, ".db") && !strings.HasSuffix(fileName, ".sql")) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的备份文件名",
		})
		return
	}

	filePath := filepath.Join(BackupDir, fileName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "备份文件不存在",
		})
		return
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除备份文件失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "备份文件删除成功",
	})
}

// 生成SQL格式的备份
func generateSQLBackup(w io.Writer) error {
	// 业务数据表列表（排除SQLite系统表）
	businessTables := []string{
		"departments",
		"employees",
		"kpi_templates",
		"kpi_items",
		"kpi_evaluations",
		"kpi_scores",
		"evaluation_comments",
		"evaluation_invitations",
		"invited_scores",
		"system_settings",
		"performance_rules",
	}

	// 写入备份头部信息
	header := fmt.Sprintf("-- KPI系统数据库备份\n-- 生成时间: %s\n-- 仅包含业务数据表\n\n",
		time.Now().Format("2006-01-02 15:04:05"))
	if _, err := w.Write([]byte(header)); err != nil {
		return err
	}

	// 为每个表生成INSERT语句
	for _, tableName := range businessTables {
		if err := generateTableBackup(w, tableName); err != nil {
			return fmt.Errorf("备份表 %s 失败: %v", tableName, err)
		}
	}

	// 写入备份尾部信息
	footer := fmt.Sprintf("\n-- 备份完成\n-- 总计 %d 个业务数据表\n", len(businessTables))
	if _, err := w.Write([]byte(footer)); err != nil {
		return err
	}

	return nil
}

// 生成单个表的备份数据
func generateTableBackup(w io.Writer, tableName string) error {
	// 写入表注释
	comment := fmt.Sprintf("\n-- 表: %s\n", tableName)
	if _, err := w.Write([]byte(comment)); err != nil {
		return err
	}

	// 生成DROP TABLE语句
	dropStmt := fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", tableName)
	if _, err := w.Write([]byte(dropStmt)); err != nil {
		return err
	}

	// 获取表结构并生成CREATE TABLE语句
	createStmt, err := generateCreateTableStatement(tableName)
	if err != nil {
		return fmt.Errorf("生成CREATE TABLE语句失败: %v", err)
	}
	if _, err := w.Write([]byte(createStmt)); err != nil {
		return err
	}

	// 查询表中的所有数据
	rows, err := models.DB.Table(tableName).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 如果表为空，跳过INSERT语句
	if !rows.Next() {
		return nil
	}

	// 准备INSERT语句前缀
	columnNames := strings.Join(columns, "`, `")
	insertPrefix := fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUES\n", tableName, columnNames)

	// 读取数据
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	first := true
	for {
		// 扫描当前行
		if err := rows.Scan(scanArgs...); err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return err
		}

		// 构建VALUES部分
		var valueStrings []string
		for _, val := range values {
			if val == nil {
				valueStrings = append(valueStrings, "NULL")
			} else {
				switch v := val.(type) {
				case string:
					// 转义字符串中的单引号
					escaped := strings.Replace(v, "'", "''", -1)
					valueStrings = append(valueStrings, fmt.Sprintf("'%s'", escaped))
				case int64:
					valueStrings = append(valueStrings, strconv.FormatInt(v, 10))
				case float64:
					valueStrings = append(valueStrings, strconv.FormatFloat(v, 'f', -1, 64))
				case bool:
					if v {
						valueStrings = append(valueStrings, "1")
					} else {
						valueStrings = append(valueStrings, "0")
					}
				case time.Time:
					// 将时间格式化为SQLite兼容的格式
					valueStrings = append(valueStrings, fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05")))
				default:
					// 其他类型转换为字符串并转义
					str := fmt.Sprintf("%v", v)
					escaped := strings.Replace(str, "'", "''", -1)
					valueStrings = append(valueStrings, fmt.Sprintf("'%s'", escaped))
				}
			}
		}

		// 写入INSERT语句
		if first {
			if _, err := w.Write([]byte(insertPrefix)); err != nil {
				return err
			}
			first = false
		} else {
			if _, err := w.Write([]byte(",\n")); err != nil {
				return err
			}
		}

		valuesStr := fmt.Sprintf("(%s)", strings.Join(valueStrings, ", "))
		if _, err := w.Write([]byte(valuesStr)); err != nil {
			return err
		}

		// 检查是否还有下一行
		if !rows.Next() {
			break
		}
	}

	// 如果有数据，写入语句结束符
	if !first {
		if _, err := w.Write([]byte(";\n")); err != nil {
			return err
		}
	}

	return nil
}

// 生成CREATE TABLE语句
func generateCreateTableStatement(tableName string) (string, error) {
	// 查询sqlite_master表获取表的创建语句
	var createSQL string
	err := models.DB.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&createSQL).Error
	if err != nil {
		return "", fmt.Errorf("查询表结构失败: %v", err)
	}

	if createSQL == "" {
		return "", fmt.Errorf("未找到表 %s 的创建语句", tableName)
	}

	// 添加分号和换行
	return createSQL + ";\n", nil
}

// 从.db文件恢复数据库
func restoreFromDBFile(backupPath string) error {
	currentDBPath := "db/kpi.db"
	backupDBPath := "db/kpi_backup_before_restore.db"

	// 创建当前数据库的备份（以防恢复失败）
	if err := copyFile(currentDBPath, backupDBPath); err != nil {
		return fmt.Errorf("创建恢复前备份失败: %v", err)
	}
	defer os.Remove(backupDBPath) // 恢复成功后删除这个备份

	// 直接替换数据库文件
	if err := copyFile(backupPath, currentDBPath); err != nil {
		// 如果恢复失败，尝试恢复原始数据库
		if restoreErr := copyFile(backupDBPath, currentDBPath); restoreErr != nil {
			return fmt.Errorf("恢复失败，且无法回滚到原始状态: %v", restoreErr)
		}
		return fmt.Errorf("恢复数据库失败，已自动回滚")
	}

	// 重新连接数据库（通过ping测试）
	if err := models.DB.Exec("SELECT 1").Error; err != nil {
		// 如果数据库连接失败，尝试恢复原始数据库
		if restoreErr := copyFile(backupDBPath, currentDBPath); restoreErr != nil {
			return fmt.Errorf("数据库连接失败，且无法回滚到原始状态: %v", restoreErr)
		}
		return fmt.Errorf("数据库恢复失败，已自动回滚")
	}

	return nil
}

// 从.sql文件恢复数据库
func restoreFromSQLFile(backupPath string) error {
	// 读取SQL文件内容
	sqlContent, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("读取SQL文件失败: %v", err)
	}

	// 创建当前数据库的备份（以防恢复失败）
	currentDBPath := "db/kpi.db"
	backupDBPath := "db/kpi_backup_before_restore.db"
	if err := copyFile(currentDBPath, backupDBPath); err != nil {
		return fmt.Errorf("创建恢复前备份失败: %v", err)
	}
	defer os.Remove(backupDBPath) // 恢复成功后删除这个备份

	// 解析SQL文件，按语句类型分组
	dropStatements, createStatements, insertStatements, err := parseSQLStatements(string(sqlContent))
	if err != nil {
		return fmt.Errorf("解析SQL语句失败: %v", err)
	}

	// 禁用外键约束（SQLite）
	if err := models.DB.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		return fmt.Errorf("禁用外键约束失败: %v", err)
	}

	// 确保恢复后重新启用外键约束
	defer func() {
		models.DB.Exec("PRAGMA foreign_keys = ON")
	}()

	// 1. 执行所有DROP TABLE语句
	for _, stmt := range dropStatements {
		if err := models.DB.Exec(stmt).Error; err != nil {
			// 重新启用外键约束
			models.DB.Exec("PRAGMA foreign_keys = ON")
			// 如果执行失败，尝试恢复原始数据库
			if restoreErr := copyFile(backupDBPath, currentDBPath); restoreErr != nil {
				return fmt.Errorf("DROP TABLE执行失败，且无法回滚到原始状态: %v", restoreErr)
			}
			return fmt.Errorf("DROP TABLE执行失败，已自动回滚: %v", err)
		}
	}

	// 2. 执行所有CREATE TABLE语句
	for _, stmt := range createStatements {
		if err := models.DB.Exec(stmt).Error; err != nil {
			// 重新启用外键约束
			models.DB.Exec("PRAGMA foreign_keys = ON")
			// 如果执行失败，尝试恢复原始数据库
			if restoreErr := copyFile(backupDBPath, currentDBPath); restoreErr != nil {
				return fmt.Errorf("CREATE TABLE执行失败，且无法回滚到原始状态: %v", restoreErr)
			}
			return fmt.Errorf("CREATE TABLE执行失败，已自动回滚: %v", err)
		}
	}

	// 3. 执行所有INSERT语句
	for _, stmt := range insertStatements {
		if err := models.DB.Exec(stmt).Error; err != nil {
			// 重新启用外键约束
			models.DB.Exec("PRAGMA foreign_keys = ON")
			// 如果执行失败，尝试恢复原始数据库
			if restoreErr := copyFile(backupDBPath, currentDBPath); restoreErr != nil {
				return fmt.Errorf("INSERT执行失败，且无法回滚到原始状态: %v", restoreErr)
			}
			return fmt.Errorf("INSERT执行失败，已自动回滚: %v", err)
		}
	}

	// 验证数据库连接
	if err := models.DB.Exec("SELECT 1").Error; err != nil {
		// 重新启用外键约束
		models.DB.Exec("PRAGMA foreign_keys = ON")
		// 如果数据库连接失败，尝试恢复原始数据库
		if restoreErr := copyFile(backupDBPath, currentDBPath); restoreErr != nil {
			return fmt.Errorf("数据库验证失败，且无法回滚到原始状态: %v", restoreErr)
		}
		return fmt.Errorf("数据库恢复失败，已自动回滚")
	}

	return nil
}

// 解析SQL语句，按类型分组
func parseSQLStatements(sqlContent string) (dropStmts, createStmts, insertStmts []string, err error) {
	// 按行分割
	lines := strings.Split(sqlContent, "\n")

	var currentStmt strings.Builder
	inMultilineComment := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 处理多行注释
		if strings.Contains(line, "/*") {
			inMultilineComment = true
		}
		if inMultilineComment {
			if strings.Contains(line, "*/") {
				inMultilineComment = false
			}
			continue
		}

		// 跳过单行注释和空行
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}

		// 累积语句
		currentStmt.WriteString(line)
		currentStmt.WriteString(" ")

		// 检查是否是完整的语句（以分号结尾）
		stmtStr := strings.TrimSpace(currentStmt.String())
		if strings.HasSuffix(stmtStr, ";") {
			// 移除末尾的分号用于分类
			cleanStmt := strings.TrimSuffix(stmtStr, ";")
			cleanStmt = strings.TrimSpace(cleanStmt)

			if cleanStmt != "" {
				upperStmt := strings.ToUpper(cleanStmt)

				// 分类语句
				if strings.HasPrefix(upperStmt, "DROP TABLE") {
					dropStmts = append(dropStmts, stmtStr)
				} else if strings.HasPrefix(upperStmt, "CREATE TABLE") {
					createStmts = append(createStmts, stmtStr)
				} else if strings.HasPrefix(upperStmt, "INSERT INTO") {
					insertStmts = append(insertStmts, stmtStr)
				}
				// 忽略其他语句（如注释等）
			}

			// 重置当前语句
			currentStmt.Reset()
		}
	}

	return dropStmts, createStmts, insertStmts, nil
}

// 辅助函数：复制文件
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
	return err
}
