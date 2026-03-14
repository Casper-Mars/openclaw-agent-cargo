package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultOpenClawDir = "~/.openclaw"
	agentsSubdir       = "agents"
	manifestFile       = "manifest.json"
	configFile         = "openclaw.json"
)

// 需要跳过的文件/目录
var skipPatterns = []string{
	".DS_Store",
	"Thumbs.db",
	".git",
	"__pycache__",
	".lock",
}

// Manifest 导出包的元信息
type Manifest struct {
	Version     string    `json:"version"`
	AgentID     string    `json:"agent_id"`
	AgentName   string    `json:"agent_name"`
	ExportTime  time.Time `json:"export_time"`
	OpenClawVer string    `json:"openclaw_version"`
	FileCount   int       `json:"file_count"`
	TotalSize   int64     `json:"total_size"`
}

// AgentConfig Agent 配置文件结构
type AgentConfig struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Model     string `json:"model"`
	Workspace string `json:"workspace"`
}

var (
	version = "1.0.0"
	verbose = false
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "openclaw-agent-cargo",
		Short: "OpenClaw Agent 导入/导出工具",
		Long:  `用于备份和恢复 OpenClaw Agent 配置的命令行工具`,
	}

	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "导出 Agent 到 tar.gz 包",
		RunE:  runExport,
	}

	var importCmd = &cobra.Command{
		Use:   "import",
		Short: "从 tar.gz 包导入 Agent",
		RunE:  runImport,
	}

	// 全局标志
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "显示详细输出")

	// export 命令标志
	exportCmd.Flags().StringP("agent", "a", "", "要导出的 Agent ID (必需)")
	exportCmd.Flags().StringP("output", "o", "", "输出文件路径 (默认：./<agent-id>-agent.tar.gz)")
	exportCmd.Flags().StringP("openclaw-dir", "d", "", "OpenClaw 目录 (默认：~/.openclaw)")
	exportCmd.MarkFlagRequired("agent")

	// import 命令标志
	importCmd.Flags().StringP("file", "f", "", "要导入的 tar.gz 文件 (必需)")
	importCmd.Flags().BoolP("force", "", false, "覆盖现有 Agent")
	importCmd.Flags().StringP("rename", "r", "", "导入时重命名 Agent")
	importCmd.Flags().StringP("openclaw-dir", "d", "", "OpenClaw 目录 (默认：~/.openclaw)")
	importCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runExport(cmd *cobra.Command, args []string) error {
	agentID, _ := cmd.Flags().GetString("agent")
	output, _ := cmd.Flags().GetString("output")
	openclawDir, _ := cmd.Flags().GetString("openclaw-dir")

	if openclawDir == "" {
		openclawDir = defaultOpenClawDir
	}
	openclawDir = expandTilde(openclawDir)

	if output == "" {
		output = fmt.Sprintf("./%s-agent.tar.gz", agentID)
	}

	agentDir := filepath.Join(openclawDir, agentsSubdir, agentID)
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		return fmt.Errorf("Agent '%s' 不存在于 %s", agentID, agentDir)
	}

	// 读取 config.json 获取 Agent 名称
	agentName := agentID
	configPath := filepath.Join(agentDir, "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg AgentConfig
		if json.Unmarshal(data, &cfg) == nil && cfg.Name != "" {
			agentName = cfg.Name
		}
	}

	// 读取 openclaw.json 获取 agent 配置
	var agentConfig map[string]interface{}
	openclawConfigPath := filepath.Join(openclawDir, configFile)
	if data, err := os.ReadFile(openclawConfigPath); err == nil {
		var fullConfig map[string]interface{}
		if json.Unmarshal(data, &fullConfig) == nil {
			// 尝试 agents.list 路径
			if agentsObj, ok := fullConfig["agents"].(map[string]interface{}); ok {
				if agentsList, ok := agentsObj["list"].([]interface{}); ok {
					for _, a := range agentsList {
						if am, ok := a.(map[string]interface{}); ok && am["id"] == agentID {
							agentConfig = am
							break
						}
					}
				}
			}
			// 尝试 list 路径（兼容旧格式）
			if agentConfig == nil {
				if agentsList, ok := fullConfig["list"].([]interface{}); ok {
					for _, a := range agentsList {
						if am, ok := a.(map[string]interface{}); ok && am["id"] == agentID {
							agentConfig = am
							break
						}
					}
				}
			}
		}
	}

	log("正在导出 Agent: %s (%s)", agentID, agentName)

	// 创建导出文件
	outFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("创建输出文件失败：%w", err)
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	var fileCount int
	var totalSize int64

	// 写入 openclaw.json 中的 agent 配置（如果有）
	if agentConfig != nil {
		configData, err := json.MarshalIndent(agentConfig, "", "  ")
		if err != nil {
			return err
		}
		header := &tar.Header{
			Name: filepath.Join(agentID, "agent-config.json"),
			Mode: 0644,
			Size: int64(len(configData)),
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tw.Write(configData); err != nil {
			return err
		}
		fileCount++
		totalSize += int64(len(configData))
		log("  ✓ agent-config.json (来自 openclaw.json)")
	}

	// 遍历 Agent 目录
	err = filepath.WalkDir(agentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过系统文件
		if shouldSkip(d.Name()) {
			return nil
		}

		// 跳过目录本身
		if d.IsDir() {
			return nil
		}

		// 获取相对路径
		relPath, err := filepath.Rel(agentDir, path)
		if err != nil {
			return err
		}

		// 读取文件信息
		info, err := d.Info()
		if err != nil {
			return err
		}

		// 创建 tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.Join(agentID, "agent-dir", relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// 写入文件内容
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := tw.Write(data); err != nil {
			return err
		}

		fileCount++
		totalSize += info.Size()
		log("  ✓ agent-dir/%s", relPath)

		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历 Agent 目录失败：%w", err)
	}

	// 检查并导出 workspace
	var workspacePath string
	if agentConfig != nil {
		if ws, ok := agentConfig["workspace"].(string); ok && ws != "" {
			workspacePath = ws
		}
	}
	
	// 如果没有配置 workspace，使用默认的全局 workspace
	if workspacePath == "" {
		workspacePath = filepath.Join(openclawDir, "workspace")
	}
	
	log("导出 workspace: %s", workspacePath)
	
	if _, err := os.Stat(workspacePath); err == nil {
		err = filepath.WalkDir(workspacePath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// 跳过系统文件
			if shouldSkip(d.Name()) {
				return nil
			}

			// 跳过目录本身
			if d.IsDir() {
				return nil
			}

			// 获取相对路径
			relPath, err := filepath.Rel(workspacePath, path)
			if err != nil {
				return err
			}

			// 读取文件信息
			info, err := d.Info()
			if err != nil {
				return err
			}

			// 创建 tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = filepath.Join(agentID, "workspace", relPath)

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			// 写入文件内容
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if _, err := tw.Write(data); err != nil {
				return err
			}

			fileCount++
			totalSize += info.Size()
			log("  ✓ workspace/%s", relPath)

			return nil
		})
		
		if err != nil {
			log("警告：遍历 workspace 失败：%v", err)
		}
	} else {
		log("警告：workspace 不存在：%s", workspacePath)
	}

	// 写入 manifest
	manifest := Manifest{
		Version:     version,
		AgentID:     agentID,
		AgentName:   agentName,
		ExportTime:  time.Now(),
		OpenClawVer: version,
		FileCount:   fileCount,
		TotalSize:   totalSize,
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name: filepath.Join(agentID, manifestFile),
		Mode: 0644,
		Size: int64(len(manifestData)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err := tw.Write(manifestData); err != nil {
		return err
	}

	log("\n导出完成!")
	log("  文件：%s", output)
	log("  文件数：%d", fileCount)
	log("  总大小：%s", formatSize(totalSize))

	return nil
}

func runImport(cmd *cobra.Command, args []string) error {
	file, _ := cmd.Flags().GetString("file")
	force, _ := cmd.Flags().GetBool("force")
	rename, _ := cmd.Flags().GetString("rename")
	openclawDir, _ := cmd.Flags().GetString("openclaw-dir")

	if openclawDir == "" {
		openclawDir = defaultOpenClawDir
	}
	openclawDir = expandTilde(openclawDir)

	// 打开导出文件
	inFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("打开文件失败：%w", err)
	}
	defer inFile.Close()

	gr, err := gzip.NewReader(inFile)
	if err != nil {
		return fmt.Errorf("解压失败：%w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	var agentID string
	var files []struct {
		header *tar.Header
		data   []byte
	}

	// 读取所有文件
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取 tar 失败：%w", err)
		}

		// 提取 agent ID（从第一个目录名）
		if agentID == "" {
			parts := strings.Split(header.Name, "/")
			if len(parts) > 0 {
				agentID = parts[0]
			}
		}

		// 跳过目录
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// 读取文件内容
		data, err := io.ReadAll(tr)
		if err != nil {
			return err
		}

		files = append(files, struct {
			header *tar.Header
			data   []byte
		}{header, data})
	}

	if agentID == "" {
		return fmt.Errorf("无法从包中提取 Agent ID")
	}

	// 应用重命名
	if rename != "" {
		log("重命名 Agent: %s -> %s", agentID, rename)
		agentID = rename
	}

	// 检查是否已存在
	targetDir := filepath.Join(openclawDir, agentsSubdir, agentID)
	if _, err := os.Stat(targetDir); err == nil {
		if !force {
			return fmt.Errorf("Agent '%s' 已存在，使用 --force 覆盖", agentID)
		}
		log("覆盖现有 Agent: %s", agentID)
	}

	log("正在导入 Agent: %s", agentID)
	log("  目标目录：%s", targetDir)

	// 创建目录并写入文件
	for _, f := range files {
		// 解析路径：agent-id/[agent-dir|workspace|agent-config.json|manifest.json]/...
		parts := strings.SplitN(f.header.Name, "/", 3)
		if len(parts) < 2 {
			continue
		}

		var targetPath string
		
		if len(parts) == 2 {
			// 根文件：agent-config.json, manifest.json
			targetPath = filepath.Join(targetDir, parts[1])
		} else {
			subdir := parts[1]
			relPath := parts[2]
			
			if subdir == "agent-dir" {
				// agent-dir 中的文件直接放到 agent 目录
				targetPath = filepath.Join(targetDir, relPath)
			} else if subdir == "workspace" {
				// workspace 文件放到全局 workspace
				targetPath = filepath.Join(openclawDir, "workspace", relPath)
			} else {
				continue
			}
		}

		// 创建父目录
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// 写入文件
		if err := os.WriteFile(targetPath, f.data, fs.FileMode(f.header.Mode)); err != nil {
			return err
		}

		log("  ✓ %s", f.header.Name)
	}

	// 更新 openclaw.json 中的 agent 配置
	agentConfigPath := filepath.Join(targetDir, "agent-config.json")
	if data, err := os.ReadFile(agentConfigPath); err == nil {
		var agentConfig map[string]interface{}
		if json.Unmarshal(data, &agentConfig) == nil {
			// 读取 openclaw.json
			openclawConfigPath := filepath.Join(openclawDir, configFile)
			var fullConfig map[string]interface{}
			if data, err := os.ReadFile(openclawConfigPath); err == nil {
				if json.Unmarshal(data, &fullConfig) == nil {
					// 尝试 agents.list 路径
					updated := false
					if agentsObj, ok := fullConfig["agents"].(map[string]interface{}); ok {
						if agentsList, ok := agentsObj["list"].([]interface{}); ok {
							for i, a := range agentsList {
								if am, ok := a.(map[string]interface{}); ok && am["id"] == agentID {
									fullConfig["agents"].(map[string]interface{})["list"].([]interface{})[i] = agentConfig
									updated = true
									break
								}
							}
							if !updated {
								fullConfig["agents"].(map[string]interface{})["list"] = append(agentsList, agentConfig)
							}
						}
					}
					// 尝试 list 路径（兼容旧格式）
					if !updated {
						if agentsList, ok := fullConfig["list"].([]interface{}); ok {
							for i, a := range agentsList {
								if am, ok := a.(map[string]interface{}); ok && am["id"] == agentID {
									fullConfig["list"].([]interface{})[i] = agentConfig
									updated = true
									break
								}
							}
							if !updated {
								fullConfig["list"] = append(agentsList, agentConfig)
							}
						}
					}

					// 写回 openclaw.json
					configData, err := json.MarshalIndent(fullConfig, "", "  ")
					if err == nil {
						if err := os.WriteFile(openclawConfigPath, configData, 0644); err == nil {
							log("  ✓ 已更新 openclaw.json")
						}
					}
				}
			}
		}
	}

	log("\n导入完成!")
	log("  Agent ID: %s", agentID)
	log("  位置：%s", targetDir)

	return nil
}



// 辅助函数

// shouldSkip 检查是否应该跳过该文件/目录
func shouldSkip(name string) bool {
	for _, pattern := range skipPatterns {
		if name == pattern || strings.HasSuffix(name, pattern) {
			return true
		}
	}
	return false
}

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

func log(format string, args ...interface{}) {
	if verbose {
		fmt.Printf(format+"\n", args...)
	}
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
