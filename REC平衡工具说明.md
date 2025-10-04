# REC 文件平衡工具说明

## 功能介绍

REC 文件平衡工具可以自动检测并补全 botmimic 数据文件夹中每个地图的 t/ct 文件夹，确保每个文件夹都有 **5 个不同的 .rec 文件**。

## 使用场景

当您批量解析多个 demo 文件后，某些地图的某些回合可能出现玩家数量不足的情况（例如某个回合的 CT 方只有 3 名玩家）。这个工具可以从同一地图的其他回合中复制 rec 文件进行补全，确保每个 t/ct 文件夹都有足够的玩家数据。

## 补全规则

1. **同地图限制**：只从同一地图文件夹下寻找文件进行补全
   - 例如：`de_dust2` 缺少文件只能从其他 `de_dust2` 的文件夹中查找
   
2. **同队伍限制**：t 文件夹只从其他 t 文件夹复制，ct 文件夹只从其他 ct 文件夹复制
   - `de_dust2/1/round0_T0-CT0/ct` 缺少文件时，可以从以下位置复制：
     - ✅ `de_dust2/1/round1_T5-CT5/ct`（同一局其他回合）
     - ✅ `de_dust2/2/round1_T5-CT5/ct`（其他局的回合）
     - ✅ `de_dust2/3/round2_T4-CT5/ct`（其他局的回合）
   - 只要不是同一个文件夹（同一局同一回合）都可以复制

3. **避免重复**：不会从同一个文件夹内复制已有的文件

4. **文件名唯一**：如果文件名冲突，会自动添加后缀（如 `player_1.rec`, `player_2.rec`）

## 使用方法

### 方法一：使用快捷脚本（推荐）

#### Windows 系统

双击运行 `rec_balance.bat` 或在命令行中执行：

```batch
rec_balance.bat [基础路径]
```

**示例：**
```batch
REM 使用默认路径
rec_balance.bat

REM 指定自定义路径
rec_balance.bat "D:\SteamLibrary\steamapps\common\Counter-Strike Global Offensive\csgo\addons\sourcemod\data\botmimic\demotest"
```

#### Linux/Mac 系统

首先给脚本添加执行权限：
```bash
chmod +x rec_balance.sh
```

然后执行：
```bash
./rec_balance.sh [基础路径]
```

**示例：**
```bash
# 使用默认路径（./output）
./rec_balance.sh

# 指定自定义路径
./rec_balance.sh /path/to/botmimic/demotest
```

### 方法二：直接使用命令行

#### 1. 编译程序

```bash
# Windows
go build -o rec_balancer.exe .\cmd\rec_balancer\main.go

# Linux/Mac
go build -o rec_balancer ./cmd/rec_balancer/main.go
```

#### 2. 执行平衡操作

```bash
# Windows
rec_balancer.exe -path="D:\路径\to\botmimic\demotest"

# Linux/Mac
./rec_balancer -path="/path/to/botmimic/demotest"
```

## 示例场景

### 场景说明

假设您有以下文件结构：

```
demotest/
└── de_dust2/
    ├── 1/
    │   ├── round0_T0-CT0/
    │   │   ├── t/
    │   │   │   ├── player1.rec
    │   │   │   ├── player2.rec
    │   │   │   └── player3.rec  (只有3个，缺2个)
    │   │   └── ct/
    │   │       ├── player4.rec
    │   │       ├── player5.rec
    │   │       └── player6.rec  (只有3个，缺2个)
    │   └── round1_T5-CT5/        ← 同一局的其他回合
    │       ├── t/
    │       │   ├── playerA.rec
    │       │   ├── playerB.rec
    │       │   ├── playerC.rec
    │       │   ├── playerD.rec
    │       │   └── playerE.rec  (有5个)
    │       └── ct/
    │           ├── playerF.rec
    │           ├── playerG.rec
    │           ├── playerH.rec
    │           ├── playerI.rec
    │           └── playerJ.rec  (有5个)
    ├── 2/
    │   └── round1_T5-CT5/
    │       ├── t/
    │       │   ├── playerK.rec
    │       │   ├── playerL.rec
    │       │   ├── playerM.rec
    │       │   ├── playerN.rec
    │       │   └── playerO.rec  (有5个)
    │       └── ct/
    │           ├── playerP.rec
    │           ├── playerQ.rec
    │           ├── playerR.rec
    │           ├── playerS.rec
    │           └── playerT.rec  (有5个)
    └── 3/
        └── round2_T4-CT4/
            ├── t/
            │   ├── playerU.rec
            │   ├── playerV.rec
            │   ├── playerW.rec
            │   └── playerX.rec  (有4个，缺1个)
            └── ct/
                ├── playerY.rec
                ├── playerZ.rec
                ├── player10.rec
                └── player11.rec  (有4个，缺1个)
```

### 执行结果

运行工具后：

```
demotest/
└── de_dust2/
    ├── 1/
    │   ├── round0_T0-CT0/
    │   │   ├── t/
    │   │   │   ├── player1.rec
    │   │   │   ├── player2.rec
    │   │   │   ├── player3.rec
    │   │   │   ├── playerA.rec  ← 从同一局的 1/round1_T5-CT5/t 复制
    │   │   │   └── playerB.rec  ← 从同一局的 1/round1_T5-CT5/t 复制
    │   │   └── ct/
    │   │       ├── player4.rec
    │   │       ├── player5.rec
    │   │       ├── player6.rec
    │   │       ├── playerF.rec  ← 从同一局的 1/round1_T5-CT5/ct 复制
    │   │       └── playerG.rec  ← 从同一局的 1/round1_T5-CT5/ct 复制
    │   └── round1_T5-CT5/  (不变，已经有5个)
    ├── 2/  (不变，已经有5个)
    └── 3/
        └── round2_T4-CT4/
            ├── t/
            │   ├── playerU.rec
            │   ├── playerV.rec
            │   ├── playerW.rec
            │   ├── playerX.rec
            │   └── playerA.rec  ← 从 1/round1_T5-CT5/t 复制
            └── ct/
                ├── playerY.rec
                ├── playerZ.rec
                ├── player10.rec
                ├── player11.rec
                └── playerF.rec  ← 从 1/round1_T5-CT5/ct 复制
```

## 工具输出

执行工具时会显示详细的处理信息：

```
========================================
REC 文件平衡工具
========================================
扫描路径: D:\...\demotest
目标数量: 每个文件夹 5 个 REC 文件
========================================

找到 3 个地图文件夹

----------------------------------------
处理地图: de_dust2
----------------------------------------
  [T] 1\round0_T0-CT0\t 缺少 2 个文件
    → 复制: playerA.rec
    → 复制: playerB.rec
    ✓ 成功补全 2 个文件
  [CT] 1\round0_T0-CT0\ct 缺少 2 个文件
    → 复制: playerF.rec
    → 复制: playerG.rec
    ✓ 成功补全 2 个文件
✓ de_dust2: 补全了 2 个文件夹

----------------------------------------
处理地图: de_mirage
----------------------------------------
✓ de_mirage: 无需补全

========================================
平衡完成！共补全 2 个文件夹
========================================
```

## 注意事项

1. **备份数据**：在首次使用前，建议备份您的原始数据
2. **目标数量**：默认每个文件夹补全到 5 个 rec 文件
3. **地图识别**：工具会自动识别以 `de_`, `cs_`, `ar_`, `aim_`, `awp_`, `fy_`, `dm_`, `surf_` 开头的文件夹作为地图文件夹
4. **文件来源**：只会从同一地图的其他文件夹复制文件
5. **不会删除**：工具只会复制文件，不会删除任何现有文件

## 推荐工作流程

1. **批量解析 demo**
   ```batch
   batch_parse.bat "path\to\demos"
   ```

2. **平衡 rec 文件**
   ```batch
   rec_balance.bat "D:\...\botmimic\demotest"
   ```

3. **在游戏中测试回放**

## 故障排除

### 问题：找不到任何地图文件夹

**解决方法：**
- 确认路径正确指向包含地图文件夹的目录
- 检查地图文件夹名称是否符合命名规范（如 `de_dust2`）

### 问题：没有找到可用的 REC 文件用于补全

**解决方法：**
- 确保同一地图下的其他文件夹中有足够的 rec 文件
- 考虑解析更多的 demo 文件来增加可用的 rec 文件数量

### 问题：文件名冲突

**解决方法：**
- 工具会自动处理文件名冲突，添加数字后缀
- 如果担心文件名混乱，可以在复制前手动重命名源文件

## 技术支持

如有问题，请查看程序输出的详细信息或联系技术支持。 