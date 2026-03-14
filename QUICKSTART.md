# 快速开始指南

## 1. 编译项目

```bash
cd openclaw-agent-cargo
make build
```

或者直接使用 Go 编译：

```bash
go build -o openclaw-agent-cargo .
```

## 2. 导出 Agent

### 基本导出

```bash
./openclaw-agent-cargo export --agent main
```

这将在当前目录创建 `main-agent.tar.gz` 文件。

### 指定输出路径

```bash
./openclaw-agent-cargo export --agent main --output ~/backups/main-backup.tar.gz
```

### 详细模式

```bash
./openclaw-agent-cargo export --agent main --verbose
```

输出示例：
```
正在导出 Agent: main (main)
  ✓ config.json
  ✓ sessions/sessions.json
  ✓ workspace/AGENTS.md
  ✓ workspace/SOUL.md
  ✓ workspace/USER.md
  ...

导出完成!
  文件：./main-agent.tar.gz
  文件数：15
  总大小：45.32 KB
```

## 3. 导入 Agent

### 基本导入

```bash
./openclaw-agent-cargo import --file main-agent.tar.gz
```

### 强制覆盖

如果目标 Agent 已存在，使用 `--force` 覆盖：

```bash
./openclaw-agent-cargo import --file main-agent.tar.gz --force
```

### 导入并重命名

```bash
./openclaw-agent-cargo import --file main-agent.tar.gz --rename main-restored
```

### 指定 OpenClaw 目录

```bash
./openclaw-agent-cargo import --file main-agent.tar.gz --openclaw-dir /path/to/.openclaw
```

## 4. 常见使用场景

### 场景 1: 备份所有 Agent

```bash
#!/bin/bash
BACKUP_DIR=~/openclaw-backups/$(date +%Y%m%d)
mkdir -p $BACKUP_DIR

# 手动列出要备份的 Agent
for agent in main developer marketing qa; do
    ./openclaw-agent-cargo export --agent $agent --output $BACKUP_DIR/$agent.tar.gz
done
```

### 场景 2: 迁移到新机器

```bash
# 在旧机器上导出
./openclaw-agent-cargo export --agent main --output main.tar.gz

# 复制文件到新机器
scp main.tar.gz user@new-machine:~/

# 在新机器上导入
./openclaw-agent-cargo import --file ~/main.tar.gz
```

### 场景 3: 创建 Agent 模板

```bash
# 导出配置完善的 Agent
./openclaw-agent-cargo export --agent main --output agent-template.tar.gz

# 之后可以多次导入作为新 Agent 的基础
./openclaw-agent-cargo import --file agent-template.tar.gz --rename new-agent-1
./openclaw-agent-cargo import --file agent-template.tar.gz --rename new-agent-2
```

### 场景 4: 恢复误删的 Agent

```bash
# 从备份恢复
./openclaw-agent-cargo import --file ~/backups/main-20260314.tar.gz --force
```

## 5. 命令参考

### export

```
导出 Agent 到 tar.gz 包

Usage:
  openclaw-agent-cargo export [flags]

Flags:
  -a, --agent string        要导出的 Agent ID (必需)
  -d, --openclaw-dir string OpenClaw 目录 (默认：~/.openclaw)
  -o, --output string       输出文件路径
  -v, --verbose             显示详细输出
```

### import

```
从 tar.gz 包导入 Agent

Usage:
  openclaw-agent-cargo import [flags]

Flags:
  -f, --file string         要导入的 tar.gz 文件 (必需)
      --force               覆盖现有 Agent
  -d, --openclaw-dir string OpenClaw 目录 (默认：~/.openclaw)
  -r, --rename string       导入时重命名 Agent
  -v, --verbose             显示详细输出
```

## 6. 故障排除

### 问题：找不到 Agent

确保 OpenClaw 目录正确：

```bash
./openclaw-agent-cargo export --agent main --openclaw-dir /custom/path/.openclaw
```

### 问题：导入时提示 Agent 已存在

使用 `--force` 覆盖，或使用 `--rename` 重命名：

```bash
./openclaw-agent-cargo import --file agent.tar.gz --force
# 或
./openclaw-agent-cargo import --file agent.tar.gz --rename new-name
```

### 问题：导出文件损坏

重新导出：

```bash
./openclaw-agent-cargo export --agent main --output agent.tar.gz --verbose
```

## 7. 下一步

- 查看 [README.md](README.md) 了解更多信息
- 查看 [examples.sh](examples.sh) 查看更多使用示例
- 提交 Issue 或 PR 到 GitHub 仓库
